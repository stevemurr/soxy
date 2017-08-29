package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"soxy/biquad/hpf"
	"soxy/biquad/lpf"
	"soxy/biquad/parametric"
	"soxy/compressor"
	"soxy/resample/smarc"

	"github.com/go-audio/audio"
	"github.com/go-audio/transforms"
	"github.com/go-audio/wav"
	"github.com/naoina/toml"
)

var (
	info     = flag.String("i", "", "get info about the file")
	inFile   = flag.String("in", "", "input file")
	inPath   = flag.String("inPath", "", "input path with many waves")
	outPath  = flag.String("outPath", "", "output folder")
	outFile  = flag.String("out", "", "target output file")
	inConfig = flag.String("c", "", "path to config")
	// norm     = flag.Bool("normalize", false, "normalize the file")
	spectro = flag.Bool("spectro", false, "also create spectrograms")
	workers = flag.Int("workers", 4, "number of go routines to use")
)

type config struct {
	Master struct {
		Gain              float64
		BitDepth          float64
		SampleRate        int
		Bandwidth         float64
		RippleFactor      float64
		RippleAttenuation float64
		Tolerance         float64
		Normalize         bool
		NormalizeTo       string
	}
	Compressor *compressor.Compressor
	Parametric []*parametric.Parametric
	HPF        *hpf.HPF
	LPF        *lpf.LPF
}

// toFloatBuffer converts the buffer to the usable format for
// processing.  We then use toIntBuffer when we want to write the
// file back to disk.
func toFloatBuffer(buf *audio.IntBuffer, bitDepth float64) *audio.FloatBuffer {
	newB := &audio.FloatBuffer{}
	newB.Data = make([]float64, len(buf.Data))
	for i := 0; i < len(buf.Data); i++ {
		newB.Data[i] = float64(buf.Data[i]) / math.Pow(2, bitDepth)
	}
	newB.Format = &audio.Format{
		NumChannels: buf.Format.NumChannels,
		SampleRate:  buf.Format.SampleRate,
	}
	return newB
}

func toIntBuffer(buf *audio.FloatBuffer, bitDepth float64) *audio.IntBuffer {
	newB := &audio.IntBuffer{}
	newB.Data = make([]int, len(buf.Data))
	for i := 0; i < len(buf.Data); i++ {
		newB.Data[i] = int(buf.Data[i] * math.Pow(2, bitDepth))
	}
	newB.Format = &audio.Format{
		NumChannels: buf.Format.NumChannels,
		SampleRate:  buf.Format.SampleRate,
	}
	return newB
}

// printInfo basic replacement for soxi - lets you peek metdata
func printInfo(name string, w *wav.Decoder) {
	fmt.Printf("Filename:\t%s\n%s:\t%d\n%s:\t%d\n%s\t%d\n", name, "NumChannels", w.Format().NumChannels, "Samplerate", w.Format().SampleRate, "Bit Depth", w.BitDepth)
}

func readConfig(inConfig string, val interface{}) error {
	fc, err := os.Open(inConfig)
	if err != nil {
		return err
	}
	defer fc.Close()
	if err := toml.NewDecoder(fc).Decode(val); err != nil {
		return err
	}
	return nil
}
func process(c config, inFile, outFile string) error {
	// fix the header preemptively
	// this is required because most of the corpus does not include a pcm chunk
	tmpFile, err := ioutil.TempFile("", "soxy")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	cmd := exec.Command("sox", inFile, "-t", "wavpcm", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(tmpFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	w := wav.NewDecoder(f)
	w.ReadInfo()

	buf, err := w.FullPCMBuffer()
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	// convert to useful floats
	buff := toFloatBuffer(buf, float64(w.BitDepth))
	// change bit depth

	transforms.Gain(buff, c.Master.Gain)

	if float64(w.BitDepth) != c.Master.BitDepth {
		transforms.Quantize(buff, 32.0)
	}
	if int(w.SampleRate) != c.Master.SampleRate {
		buff.Data = smarc.Resample(buff.Data, int(w.SampleRate), c.Master.SampleRate, c.Master.Bandwidth, c.Master.RippleFactor, c.Master.RippleAttenuation, c.Master.Tolerance)
	}
	// eq and compression
	if c.HPF != nil {
		hpf.HighPass(buff, c.HPF.Freq, float64(w.SampleRate), int(w.NumChans))
	}
	if c.LPF != nil {
		lpf.LowPass(buff, c.LPF.Freq, float64(w.SampleRate), int(w.NumChans))
	}
	if len(c.Parametric) != 0 {
		for _, eq := range c.Parametric {
			parametric.EQ(buff, eq.Freq, eq.Gain, eq.Q, float64(c.Master.SampleRate), int(w.NumChans))
		}
	}
	if c.Compressor != nil {
		compressor.Compress(buff, c.Compressor.Ratio, c.Compressor.AttackTime, c.Compressor.ReleaseTime, c.Compressor.Threshold, c.Compressor.InputGain, c.Compressor.OutputGain, float64(w.SampleRate), c.Compressor.LookAheadDelay, c.Compressor.Knee)
	}

	if *spectro {
		defer func() {
			cmd := exec.Command("sox", out.Name(), "-n", "spectrogram", "-o", "spectro_"+out.Name())
			cmd.Run()
		}()
	}

	if c.Master.Normalize {
		// reuse the tmpfile
		wr := wav.NewEncoder(tmpFile, c.Master.SampleRate, int(c.Master.BitDepth), int(w.NumChans), int(w.WavAudioFormat))
		if err := wr.Write(toIntBuffer(buff, float64(c.Master.BitDepth))); err != nil {
			log.Fatal(err)
		}
		if err = wr.Close(); err != nil {
			log.Fatal(err)
		}
		// use sox to normalize
		cmd := exec.Command("sox", tmpFile.Name(), out.Name(), "--norm="+c.Master.NormalizeTo)
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
		return nil
	}
	// write the file out
	wr := wav.NewEncoder(out, c.Master.SampleRate, int(c.Master.BitDepth), int(w.NumChans), int(w.WavAudioFormat))
	if err := wr.Write(toIntBuffer(buff, float64(c.Master.BitDepth))); err != nil {
		log.Fatal(err)
	}
	if err = wr.Close(); err != nil {
		log.Fatal(err)
	}
	return nil
}

type job struct {
	InFile  string
	OutFile string
	C       config
}

// worker consumes the jobs channel
func worker(jobs <-chan job, results chan<- string) {
	for j := range jobs {
		if err := process(j.C, j.InFile, j.OutFile); err != nil {
			results <- fmt.Sprintf("!!! %s failed\n", j.InFile)
		}
		results <- fmt.Sprintf("%s done ...\n", j.InFile)
	}
}

func main() {
	flag.Parse()

	if *info != "" {
		f, err := os.Open(*info)
		if err != nil {
			log.Fatal(err)
		}
		w := wav.NewDecoder(f)
		w.ReadInfo()
		_, name := filepath.Split(*info)
		printInfo(name, w)
		os.Exit(0)
	}
	var c config
	if err := readConfig(*inConfig, &c); err != nil {
		panic(err)
	}
	files, err := filepath.Glob(filepath.Join(*inPath, "*.wav"))
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(*outPath, 0755)
	jobs := make(chan job, len(files))
	results := make(chan string, len(files))

	// start the pool
	for idx := 0; idx < *workers; idx++ {
		go worker(jobs, results)
	}
	for _, fi := range files {
		_, tail := filepath.Split(fi)
		pOut := filepath.Join(*outPath, tail)
		jobs <- job{InFile: fi, OutFile: pOut, C: c}
	}
	close(jobs)

	for range files {
		fmt.Printf("%s", <-results)
	}
}

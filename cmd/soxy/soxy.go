package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"soxy/biquad/hpf"
	"soxy/biquad/lpf"
	"soxy/biquad/parametric"
	"soxy/compressor"
	"soxy/resample/smarc"
	"soxy/tempr"
	"strconv"

	"github.com/go-audio/audio"
	"github.com/go-audio/transforms"
	"github.com/go-audio/wav"
	"github.com/naoina/toml"
	"gopkg.in/cheggaaa/pb.v1"
)

var bitDepthConvert = map[float64]string{
	24.0: "pcm_s24le",
	16.0: "pcm_s16le",
}

var (
	info     = flag.String("i", "", "get info about the file")
	inPath   = flag.String("inPath", "", "input path with many waves")
	outPath  = flag.String("outPath", "", "output folder")
	inConfig = flag.String("c", "", "path to config")
	spectro  = flag.Bool("spectro", false, "also create spectrograms")
	workers  = flag.Int("workers", runtime.NumCPU(), "Number of go routines to use.")
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

		IntegratedLoudness string
		LoudnessRange      string
		TruePeak           string
		PeakNorm           string

		SoxNorm   bool
		SoxNormTo string
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
	// convert to float buffer with range -1 to 1
	buff := toFloatBuffer(buf, float64(w.BitDepth))
	transforms.Gain(buff, c.Master.Gain)

	// Resample to 192000 for internal processing.
	buff.Data = smarc.Resample(buff.Data, int(w.SampleRate), 192000, c.Master.Bandwidth, c.Master.RippleFactor, c.Master.RippleAttenuation, c.Master.Tolerance)

	if c.HPF != nil {
		hpf.HighPass(buff, c.HPF.Freq, 192000.0, int(w.NumChans))
	}
	if c.LPF != nil {
		lpf.LowPass(buff, c.LPF.Freq, 192000.0, int(w.NumChans))
	}
	if len(c.Parametric) != 0 {
		for _, eq := range c.Parametric {
			parametric.EQ(buff, eq.Freq, eq.Gain, eq.Q, 192000.0, int(w.NumChans))
		}
	}
	if c.Compressor != nil {
		compressor.Compress(buff, c.Compressor.Ratio, c.Compressor.AttackTime, c.Compressor.ReleaseTime, c.Compressor.Threshold, c.Compressor.InputGain, c.Compressor.OutputGain, 192000, c.Compressor.LookAheadDelay, c.Compressor.Knee)
	}

	if *spectro {
		// dump metrics and stats in output folder
		defer func() {
			// draw spectrogram
			specFol := filepath.Join(*outPath, "Spectrograms")
			os.MkdirAll(specFol, 0755)
			_, tail := filepath.Split(out.Name())
			pngFile := filepath.Join(specFol, tail[:len(tail)-4]+".png")
			cmd := exec.Command("sox", out.Name(), "-n", "spectrogram", "-o", pngFile)
			cmd.Run()

			// draw the waveform
			dur, err := w.Duration()
			if err != nil {
				log.Fatal(err)
			}
			waveFol := filepath.Join(*outPath, "Waveforms")
			os.MkdirAll(waveFol, 0755)
			wfFile := filepath.Join(waveFol, tail[:len(tail)-4]+".png")
			dura := dur.String()
			cmd = exec.Command("audiowaveform", "-i", out.Name(), "-o", wfFile, "-b", "16", "-e", dura[:len(dura)-1])
			cmd.Run()

			// save the config
			ff, err := ioutil.ReadFile(*inConfig)
			if err != nil {
				log.Fatal(err)
			}
			configFol := filepath.Join(*outPath, "Config")
			os.MkdirAll(configFol, 0755)
			_, ctail := filepath.Split(*inConfig)
			if err := ioutil.WriteFile(filepath.Join(configFol, ctail), ff, 0644); err != nil {
				log.Fatal(err)
			}
			statsFol := filepath.Join(*outPath, "Stats")
			os.MkdirAll(statsFol, 0755)

			var cmdBuf bytes.Buffer
			cmd = exec.Command("sox", out.Name(), "-n", "stats")
			cmd.Stdout = &cmdBuf
			cmd.Stderr = &cmdBuf
			cmd.Run()
			if err := ioutil.WriteFile(filepath.Join(statsFol, tail[:len(tail)-4]+".txt"), cmdBuf.Bytes(), 0644); err != nil {
				panic(err)
			}
		}()
	}

	// write the file down as 192
	wr := wav.NewEncoder(tmpFile, 192000.0, int(w.BitDepth), int(w.NumChans), int(w.WavAudioFormat))
	if err := wr.Write(toIntBuffer(buff, float64(w.BitDepth))); err != nil {
		panic(err)
	}
	if err = wr.Close(); err != nil {
		panic(err)
	}
	newRate := strconv.Itoa(c.Master.SampleRate)
	if c.Master.Normalize {
		// Loudness normalization first
		normTemp, err := tempr.TempFile("", "soxy", ".wav")
		if err != nil {
			panic(err)
		}
		loudNormSettings := fmt.Sprintf("loudnorm=I=%s:LRA=%s:TP=%s", c.Master.IntegratedLoudness, c.Master.LoudnessRange, c.Master.TruePeak)
		command := []string{
			"-y",
			"-i",
			tmpFile.Name(),
			"-af",
			loudNormSettings,
			"-acodec",
			bitDepthConvert[c.Master.BitDepth],
			"-ar",
			newRate,
			normTemp.Name(),
		}
		cmd := exec.Command("ffmpeg", command...)
		if err := cmd.Run(); err != nil {
			panic(err)
		}
		// Peak normalization
		cmd = exec.Command("sox", normTemp.Name(), out.Name(), "--norm="+c.Master.PeakNorm)
		cmd.Run()
		return nil
	}
	// just do a conversion
	cmd = exec.Command("ffmpeg", "-y", "-i", tmpFile.Name(), "-acodec", bitDepthConvert[c.Master.BitDepth], "-ar", newRate, out.Name())
	cmd.Run()
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
	// make output path - delete if exists
	// if err := os.MkdirAll(*outPath, 0755); err != nil {
	// 	os.RemoveAll(*outPath)
	// 	os.MkdirAll(*outPath, 0755)
	// }
	os.RemoveAll(*outPath)
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

	bar := pb.StartNew(len(files))
	for range files {
		bar.Increment()
		<-results
	}
	bar.Finish()
}

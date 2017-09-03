# Audio Tool

This is an advanced version of SOX written in Go with all the algorithms implemented from scratch save for the high quality resampling.

# Installation

For development you will need Go, SOX and (audiowaveform)[https://github.com/bbc/audiowaveform]

# Usage

Pass an in file, out file and config file to the binary.

`main -c config/uprez.toml -in path/to/file.wav -out path/to/outfile.wav`

# Config file

The config file describes the types of transforms to apply to the audio.  The types of transforms are:
1) Bit depth conversion (i.e. 16 -> 24)
2) Resampling (i.e. 44.1k -> 48k)
3) High and Low pass filters
4) Constant Q parametric
5) Variable-knee lookahead compressor

# Example config

```toml
# Master section lets you define target sample rate and bit depth.
[master]
# Scale input before processing.
gain=0.8
# Target bit depth
bitdepth=24.0
# Target sample rate
samplerate=48000
# 0.95 - 0.99 - may fail above .97
bandwidth=0.97
# 0.01 - 0.1 - may fail the lower you go
ripplefactor=0.1
# 100 - 159
rippleattenuation=150.0
# Don't edit
tolerance=0.000001

# Compressor lets you apply compression :)
[compressor]
inputgain=1.0
outputgain=1.0
threshold=-5.0
# 30 - 300 but can handle any input techinally.
# For example you can use a super fast attack like 0.1
attacktime=0.1
releasetime=150.0
# 1 - 20
ratio=2.0
# 0 - 20.  0 = hard knee and 20 = soft knee
knee=10.0
# How much look ahead time.  If many transients this can solve
# the slow compressor problem
lookaheaddelay=5000.0
# No use
stereolink=0
# No Use
processortype=0
# Should be same as samplerate in master section
samplerate=48000.0
# "analog" setting.  This effects how the attack and release 
# are calculated.  See compressor module for algorithm.
analog=false

# HPF and LPF filters. 
# Butterworth curve 
[hpf]
freq=0.0
[lpf]
freq=20000.0

# Peaking filters - using constant q so these sound reasonable 
# at various gain settings.
[[parametric]]
freq=200.0
gain=1.0
q=0.5
[[parametric]]
freq=5000.0
gain=3.0
q=1.0
[[parametric]]
freq=15000.0
gain=8.0
q=1.0
# [[parametric]]
# freq=20000.0
# gain=3.0
# q=0.3
```

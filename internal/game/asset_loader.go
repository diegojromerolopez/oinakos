package game

import (
	"image"
	"io/fs"
	"runtime"
	"sync"

	"oinakos/internal/engine"
)

type SpriteLoadJob struct {
	Path    string
	Dest    *interface{}
	Decoded image.Image
}

// loadSpritesParallel performs background PNG decoding and then finalizes Ebiten images on the main thread
func loadSpritesParallel(assets fs.FS, jobs []*SpriteLoadJob, graphics engine.Graphics) {
	workerCount := runtime.NumCPU()
	if workerCount < 4 {
		workerCount = 4
	}

	jobChan := make(chan *SpriteLoadJob, len(jobs))
	for _, j := range jobs {
		jobChan <- j
	}
	close(jobChan)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				img, err := engine.DecodeSpriteRaw(assets, job.Path, true)
				if err == nil && img != nil {
					job.Decoded = img
				}
			}
		}()
	}
	wg.Wait()

	// In the main thread, upload to GPU and assign pointers
	for _, job := range jobs {
		if job.Decoded != nil {
			*job.Dest = graphics.NewImageFromImage(job.Decoded)
		}
	}
}

type AudioLoadJob struct {
	Name string
	Path string
	Data []byte
}

// loadAudioParallel performs background WAV/MP3 decoding
func loadAudioParallel(assets fs.FS, jobs []*AudioLoadJob) {
	workerCount := runtime.NumCPU()
	if workerCount < 4 {
		workerCount = 4
	}

	jobChan := make(chan *AudioLoadJob, len(jobs))
	for _, j := range jobs {
		if j != nil {
			jobChan <- j
		}
	}
	close(jobChan)

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				data, err := engine.DecodeAudioRaw(assets, job.Path)
				if err == nil {
					job.Data = data
				}
			}
		}()
	}
	wg.Wait()

	// Register in GlobalAudio (main thread for safety, though LoadSound is now thread-safe)
	if engine.GlobalAudio != nil {
		for _, job := range jobs {
			if job != nil && job.Data != nil {
				engine.GlobalAudio.LoadSound(job.Name, job.Path)
			}
		}
	}
}

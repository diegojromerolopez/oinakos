package game

import (
	"image"
	"io/fs"
	"log"
	"runtime"
	"sync"
	"sync/atomic"

	"oinakos/internal/engine"
)

type SpriteLoadJob struct {
	Path    string
	Dest    *interface{}
	Decoded image.Image
}

// loadSpritesParallel performs background PNG decoding and then finalizes Ebiten images on the main thread
func loadSpritesParallel(assets fs.FS, jobs []*SpriteLoadJob, graphics engine.Graphics, progress *int32) {
	if len(jobs) == 0 {
		if progress != nil { atomic.StoreInt32(progress, 1000) }
		return
	}

	workerCount := runtime.NumCPU()
	if runtime.GOOS == "js" {
		workerCount = 2
	} else if workerCount < 4 {
		workerCount = 4
	}

	jobChan := make(chan *SpriteLoadJob, len(jobs))
	for _, j := range jobs {
		jobChan <- j
	}
	close(jobChan)

	var wg sync.WaitGroup
	var completed int32
	total := int32(len(jobs))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				img, err := engine.DecodeSpriteRaw(assets, job.Path, true)
				if err == nil && img != nil {
					job.Decoded = img
				}
				
				if progress != nil {
					done := atomic.AddInt32(&completed, 1)
					// Reserve 0-800 for decoding, 800-1000 for GPU upload
					p := int32(float64(done) / float64(total) * 800)
					atomic.StoreInt32(progress, p)
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()

	// In the main thread, upload to GPU and assign pointers
	for i, job := range jobs {
		if job.Decoded != nil {
			*job.Dest = graphics.NewImageFromImage(job.Decoded)
		}
		if progress != nil {
			p := 800 + int32((float64(i+1)/float64(total))*200)
			atomic.StoreInt32(progress, p)
		}
		if i%10 == 0 { runtime.Gosched() }
	}
}

type AudioLoadJob struct {
	Name string
	Path string
	Data []byte
}

// loadAudioParallel performs background WAV/MP3 decoding
func loadAudioParallel(assets fs.FS, jobs []*AudioLoadJob, progress *int32) {
	if len(jobs) == 0 {
		if progress != nil { atomic.StoreInt32(progress, 1000) }
		return
	}

	workerCount := runtime.NumCPU()
	if runtime.GOOS == "js" {
		workerCount = 2
	} else if workerCount < 4 {
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
	var completed int32
	total := int32(len(jobs))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				log.Printf("Worker decoding %s (%s)...", job.Name, job.Path)
				data, err := engine.DecodeAudioRaw(assets, job.Path)
				if err == nil {
					job.Data = data
				}
				if progress != nil {
					done := atomic.AddInt32(&completed, 1)
					p := int32(float64(done) / float64(total) * 1000)
					atomic.StoreInt32(progress, p)
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()

	// Register in GlobalAudio (using pre-decoded bytes to avoid double-decoding)
	if engine.GlobalAudio != nil {
		for _, job := range jobs {
			if job != nil && job.Data != nil {
				engine.GlobalAudio.LoadSoundFromBytes(job.Name, job.Data)
			}
		}
	}
	if progress != nil { atomic.StoreInt32(progress, 1000) }
}

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
	Dest    *engine.Image
	Decoded image.Image
}

// LoadingState tracks progress across multiple asynchronous loading tasks (images, audio).
type LoadingState struct {
	Total    int32
	Finished int32
	Progress *int32
}

func (ls *LoadingState) Add(count int) {
	if ls == nil { return }
	atomic.AddInt32(&ls.Total, int32(count))
}

func (ls *LoadingState) Done() {
	if ls == nil { return }
	finished := atomic.AddInt32(&ls.Finished, 1)
	total := atomic.LoadInt32(&ls.Total)
	if total > 0 {
		atomic.StoreInt32(ls.Progress, (finished*1000)/total)
	}
}

func loadSpritesParallel(assets fs.FS, jobs []*SpriteLoadJob, graphics engine.Graphics, ls *LoadingState) {
	if len(jobs) == 0 {
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

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				img, err := engine.DecodeSpriteRaw(assets, job.Path, true)
				if err == nil && img != nil {
					job.Decoded = img
				}
				
				if ls != nil {
					ls.Done()
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
		if ls != nil {
			ls.Done()
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
func loadAudioParallel(assets fs.FS, jobs []*AudioLoadJob, ls *LoadingState) {
	if len(jobs) == 0 {
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
				if ls != nil {
					ls.Done()
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()

	// Register in GlobalAudio (using pre-decoded bytes to avoid double-decoding)
	if engine.GlobalAudio != nil {
		for i, job := range jobs {
			if job != nil && job.Data != nil {
				engine.GlobalAudio.LoadSoundFromBytes(job.Name, job.Data)
			}
			if ls != nil {
				ls.Done()
			}
			if i%10 == 0 { runtime.Gosched() }
		}
	}
}

package banana

import (
	"io"
	"log"

	"github.com/fsnotify/fsnotify"
)

type ChangeListener interface {
	OnChange() error
}

func StartWatching(dirs []string, l ChangeListener) (io.Closer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	refresh := make(chan bool, 1)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				evs := fsnotify.Write | fsnotify.Create | fsnotify.Remove
				if event.Op&evs > 0 {
					select {
					case refresh <- true:
					default:
						// If we're already waiting to refresh, ignore
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	go func() {
		for _ = range refresh {
			err := l.OnChange()
			if err != nil {
				log.Println(err)
			}
		}
	}()

	for _, dir := range dirs {
		err = watcher.Add(dir)
		if err != nil {
			return nil, err
		}
	}

	return watcher, nil
}

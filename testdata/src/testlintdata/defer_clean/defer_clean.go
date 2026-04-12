package defer_clean

import (
    "os"
    "sync"
)

func badLock(p *sync.Mutex) int {
    p.Lock()
    if p == nil {
        p.Unlock() // want "use defer to clean up resources such as files and locks"
        return 0
    }
    p.Unlock() // want "use defer to clean up resources such as files and locks"
    return 1
}

func goodLock(p *sync.Mutex) int {
    p.Lock()
    defer p.Unlock()
    if p == nil {
        return 0
    }
    return 1
}

func badFile() error {
    f, _ := os.Open("file")
    _ = f
    f.Close() // want "use defer to clean up resources such as files and locks"
    return nil
}

func goodFile() error {
    f, _ := os.Open("file")
    defer f.Close()
    _ = f
    return nil
}

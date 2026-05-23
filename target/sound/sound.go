package sound

import (
    "crypto/sha1"
    "encoding/binary"
    "io"
    "math"
    "math/rand"
    "net/http"
    "sync"
)

var mu sync.RWMutex
var buffers map[string][]byte = make(map[string][]byte)

var srGlobal = 48000
var secondsGlobal = 5
var channelsGlobal = 2
var bitsPerSampleGlobal = 24
var maxBuffersGlobal = 1000

func Init(sr int, seconds int, channels int, bitsPerSample int, maxBuffers int) {
    if sr > 0 {
        srGlobal = sr
    }
    if seconds > 0 {
        secondsGlobal = seconds
    }
    if channels > 0 {
        channelsGlobal = channels
    }
    if bitsPerSample > 0 {
        bitsPerSampleGlobal = bitsPerSample
    }
    if maxBuffers > 0 {
        maxBuffersGlobal = maxBuffers
    }
}

func SynthesizeAndStore(id string) {
    buf := synthScream(id, srGlobal, secondsGlobal, channelsGlobal, bitsPerSampleGlobal)
    mu.Lock()
    buffers[id] = buf
    if len(buffers) > maxBuffersGlobal {
        for k := range buffers {
            delete(buffers, k)
            break
        }
    }
    mu.Unlock()
}

func GetBuffer(id string) ([]byte, bool) {
    mu.RLock()
    b, ok := buffers[id]
    mu.RUnlock()
    return b, ok
}

// Handler for /scream/{id}
func Handler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Path[len("/scream/"):]
    if id == "" {
        http.Error(w, "missing id", http.StatusBadRequest)
        return
    }
    b, ok := GetBuffer(id)
    if !ok {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "audio/wav")
    if err := writeWavHeader(w, len(b), srGlobal, channelsGlobal, bitsPerSampleGlobal); err != nil {
        http.Error(w, "internal", http.StatusInternalServerError)
        return
    }
    _, _ = w.Write(b)
}

func synthScream(seed string, sr int, seconds int, channels int, bitsPerSample int) []byte {
    if channels < 1 {
        channels = 1
    }
    if bitsPerSample != 16 && bitsPerSample != 24 && bitsPerSample != 32 {
        bitsPerSample = 24
    }

    n := sr * seconds
    bytesPerSample := bitsPerSample / 8
    frameBytes := channels * bytesPerSample
    buf := make([]byte, 0, n*frameBytes)

    h := sha1.Sum([]byte(seed))
    fbase := 300.0 + float64((int(h[0])<<8|int(h[1]))%1800)
    randSeed := int64(binaryLittleUint64(h[:8]))
    rnd := rand.New(rand.NewSource(randSeed))

    for i := 0; i < n; i++ {
        t := float64(i) / float64(sr)
        env := math.Exp(-3.0 * t / float64(seconds))
        freq := fbase * (1.0 + 0.2*math.Sin(2*math.Pi*2.0*t))
        s := math.Sin(2*math.Pi*freq*t)
        noise := (rnd.Float64()*2.0 - 1.0) * 0.3
        sample := 0.7*s*env + 0.3*noise*env
        left := int32(math.Max(-8388607, math.Min(8388607, sample*8388607)))
        right := int32(math.Max(-8388607, math.Min(8388607, (sample*0.97+0.03*math.Sin(2*math.Pi*freq*1.003*t))*8388607)))

        for ch := 0; ch < channels; ch++ {
            value := left
            if ch%2 == 1 {
                value = right
            }
            switch bitsPerSample {
            case 16:
                buf = append(buf, byte(value), byte(value>>8))
            case 24:
                buf = append(buf, byte(value), byte(value>>8), byte(value>>16))
            case 32:
                buf = append(buf, byte(value), byte(value>>8), byte(value>>16), byte(value>>24))
            }
        }
    }
    return buf
}

func binaryLittleUint64(b []byte) uint64 {
    var v uint64
    for i := 0; i < len(b) && i < 8; i++ {
        v |= uint64(b[i]) << (8 * uint(i))
    }
    return v
}

func writeWavHeader(w io.Writer, dataLen int, sr int, channels int, bitsPerSample int) error {
    if channels < 1 {
        channels = 1
    }
    if bitsPerSample != 16 && bitsPerSample != 24 && bitsPerSample != 32 {
        bitsPerSample = 24
    }

    var subchunk2Size = uint32(dataLen)
    var chunkSize = uint32(36) + subchunk2Size
    if _, err := w.Write([]byte("RIFF")); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, chunkSize); err != nil {
        return err
    }
    if _, err := w.Write([]byte("WAVE")); err != nil {
        return err
    }
    if _, err := w.Write([]byte("fmt ")); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, uint32(16)); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, uint16(1)); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, uint16(channels)); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, uint32(sr)); err != nil {
        return err
    }
    byteRate := uint32(sr * channels * bitsPerSample / 8)
    if err := binary.Write(w, binary.LittleEndian, byteRate); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, uint16(channels*bitsPerSample/8)); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
        return err
    }
    if _, err := w.Write([]byte("data")); err != nil {
        return err
    }
    if err := binary.Write(w, binary.LittleEndian, subchunk2Size); err != nil {
        return err
    }
    return nil
}

package log

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duanhf2012/origin/v2/util/bytespool"
)

var memPool = bytespool.NewMemAreaPool()

type IoWriter struct {
	outFile    io.Writer // destination for output
	writeBytes int64
	logChannel chan []byte
	wg         sync.WaitGroup
	closeSig   chan struct{}

	lockWrite sync.Mutex

	filePath       string
	filePrefix     string
	fileDay        int
	fileCreateTime int64 //second
}

func (iw *IoWriter) Close() error {
	iw.lockWrite.Lock()
	defer iw.lockWrite.Unlock()

	iw.close()

	return nil
}

func (iw *IoWriter) close() error {
	if iw.closeSig != nil {
		close(iw.closeSig)
		iw.closeSig = nil
	}
	iw.wg.Wait()

	if iw.outFile != nil {
		err := iw.outFile.(io.Closer).Close()
		iw.outFile = nil
		return err
	}

	return nil
}

func (iw *IoWriter) writeFile(p []byte) (n int, err error) {
	//switch log file
	iw.switchFile()

	if iw.outFile != nil {
		n, _ = iw.outFile.Write(p)
		if n > 0 {
			atomic.AddInt64(&iw.writeBytes, int64(n))
		}
	}

	return 0, nil
}

func (iw *IoWriter) Write(p []byte) (n int, err error) {
	iw.lockWrite.Lock()
	defer iw.lockWrite.Unlock()

	if iw.logChannel == nil {
		return iw.writeIo(p)
	}

	copyBuff := memPool.MakeBytes(len(p))
	if copyBuff == nil {
		return 0, fmt.Errorf("MakeByteSlice failed")
	}
	copy(copyBuff, p)

	iw.logChannel <- copyBuff

	return
}

func (iw *IoWriter) writeIo(p []byte) (n int, err error) {
	n, err = iw.writeFile(p)

	if OpenConsole {
		n, err = os.Stdout.Write(p)
	}

	return
}

func (iw *IoWriter) setLogChannel(logChannelNum int) (err error) {
	iw.lockWrite.Lock()
	defer iw.lockWrite.Unlock()
	iw.close()

	if logChannelNum == 0 {
		return nil
	}

	//copy iw.logChannel
	var logInfo []byte
	logChannel := make(chan []byte, logChannelNum)
	for i := 0; i < logChannelNum && i < len(iw.logChannel); i++ {
		logInfo = <-iw.logChannel
		logChannel <- logInfo
	}
	iw.logChannel = logChannel

	iw.closeSig = make(chan struct{})
	iw.wg.Add(1)
	go iw.run()

	return nil
}

func (iw *IoWriter) run() {
	defer iw.wg.Done()

Loop:
	for {
		select {
		case <-iw.closeSig:
			break Loop
		case logs := <-iw.logChannel:
			iw.writeIo(logs)
			memPool.ReleaseBytes(logs)
		}
	}

	for len(iw.logChannel) > 0 {
		logs := <-iw.logChannel
		iw.writeIo(logs)
		memPool.ReleaseBytes(logs)
	}
}

func (iw *IoWriter) isFull() bool {
	if LogSize == 0 {
		return false
	}

	return atomic.LoadInt64(&iw.writeBytes) >= LogSize
}

func (iw *IoWriter) switchFile() error {
	now := time.Now()
	if iw.fileCreateTime == now.Unix() {
		return nil
	}

	if iw.fileDay == now.Day() && !iw.isFull() {
		return nil
	}

	if iw.filePath != "" {
		var err error
		fileName := fmt.Sprintf("%s%d%02d%02d_%02d_%02d_%02d.log",
			iw.filePrefix,
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		filePath := path.Join(iw.filePath, fileName)

		iw.outFile, err = os.Create(filePath)
		if err != nil {
			return err
		}
		iw.fileDay = now.Day()
		iw.fileCreateTime = now.Unix()
		atomic.StoreInt64(&iw.writeBytes, 0)
	}

	return nil
}

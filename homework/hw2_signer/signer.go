package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

/*
Само задание по сути состоит из двух частей
* Написание функции ExecutePipeline которая обеспечивает нам конвейерную обработку функций-воркеров, которые что-то делают.
* Написание нескольких функций, которые считают нам какую-то условную хеш-сумму от входных данных

Расчет хеш-суммы реализован следующей цепочкой:
* SingleHash считает значение crc32(data)+"~"+crc32(md5(data)) ( конкатенация двух строк через ~),
где data - то что пришло на вход (по сути - числа из первой функции)
* MultiHash считает значение crc32(th+data)) (конкатенация цифры, приведённой
	 к строке и строки), где th=0..5 ( т.е. 6 хешей на каждое входящее значение ),
потом берёт конкатенацию результатов в порядке расчета (0..5), где data - то что пришло
на вход (и ушло на выход из SingleHash)

* crc32 считается через функцию DataSignerCrc32
* md5 считается через DataSignerMd5
*/

// сюда писать код

//ExecutePipeline
/*
type job func(in, out chan interface{})

var (
	dataSignerOverheat uint32 = 0
	DataSignerSalt            = ""
)

var OverheatLock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
			fmt.Println("OverheatLock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var OverheatUnlock = func() {
	for {
		if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
			fmt.Println("OverheatUnlock happend")
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}

var DataSignerMd5 = func(data string) string {
	OverheatLock()
	defer OverheatUnlock()
	data += DataSignerSalt
	dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
	time.Sleep(10 * time.Millisecond)
	return dataHash
}

var DataSignerCrc32 = func(data string) string {
	data += DataSignerSalt
	crcH := crc32.ChecksumIEEE([]byte(data))
	dataHash := strconv.FormatUint(uint64(crcH), 10)
	time.Sleep(time.Second)
	return dataHash
}

func main() {

	testExpected := "1173136728138862632818075107442090076184424490584241521304_1696913515191343735512658979631549563179965036907783101867_27225454331033649287118297354036464389062965355426795162684_29568666068035183841425683795340791879727309630931025356555_3994492081516972096677631278379039212655368881548151736_4958044192186797981418233587017209679042592862002427381542_4958044192186797981418233587017209679042592862002427381542"
	testResult := "NOT_SET"

	// это небольшая защита от попыток не вызывать мои функции расчета
	// я преопределяю фукции на свои которые инкрементят локальный счетчик
	// переопределение возможо потому что я объявил функцию как переменную, в которой лежит функция
	var (
		DataSignerSalt         string = "" // на сервере будет другое значение
		OverheatLockCounter    uint32
		OverheatUnlockCounter  uint32
		DataSignerMd5Counter   uint32
		DataSignerCrc32Counter uint32
	)
	OverheatLock = func() {
		atomic.AddUint32(&OverheatLockCounter, 1)
		for {
			if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 0, 1); !swapped {
				fmt.Println("OverheatLock happend")
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	OverheatUnlock = func() {
		atomic.AddUint32(&OverheatUnlockCounter, 1)
		for {
			if swapped := atomic.CompareAndSwapUint32(&dataSignerOverheat, 1, 0); !swapped {
				fmt.Println("OverheatUnlock happend")
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	DataSignerMd5 = func(data string) string {
		atomic.AddUint32(&DataSignerMd5Counter, 1)
		OverheatLock()
		defer OverheatUnlock()
		data += DataSignerSalt
		dataHash := fmt.Sprintf("%x", md5.Sum([]byte(data)))
		time.Sleep(10 * time.Millisecond)
		return dataHash
	}
	DataSignerCrc32 = func(data string) string {
		atomic.AddUint32(&DataSignerCrc32Counter, 1)
		data += DataSignerSalt
		crcH := crc32.ChecksumIEEE([]byte(data))
		dataHash := strconv.FormatUint(uint64(crcH), 10)
		time.Sleep(time.Second)
		return dataHash
	}

	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	//inputData := []int{0, 1}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			fmt.Println(dataRaw)
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			testResult = data
		}),
	}

	start := time.Now()

	ExecutePipeline(hashSignJobs...)
	//ExecutePipeline(freeFlowJobs...)

	end := time.Since(start)

	expectedTime := 3 * time.Second

	fmt.Println("\nexecution time: %s\n", end)

	if testExpected != testResult {
		fmt.Println("results not match\nGot: %v\nExpected: %v", testResult, testExpected)
	}

	if end > expectedTime {
		fmt.Println("execition too long\nGot: %s\nExpected: <%s", end, time.Second*3)
	}

	// 8 потому что 2 в SingleHash и 6 в MultiHash
	if int(OverheatLockCounter) != len(inputData) ||
		int(OverheatUnlockCounter) != len(inputData) ||
		int(DataSignerMd5Counter) != len(inputData) ||
		int(DataSignerCrc32Counter) != len(inputData)*8 {
		fmt.Println("not enough hash-func calls")
	}

	start = time.Now()

	ExecutePipeline(freeFlowJobs...)

	end = time.Since(start)

	fmt.Println("\nexecution time: %s\n", end)
} */

var SingleHash = job(func(in, out chan interface{}) {

	var recieved uint32

	mu := &sync.Mutex{}
	for i := range in {

		//recieved++
		atomic.AddUint32(&recieved, 1)
		//fmt.Println("\tget", i)
		go func(inString string, mu *sync.Mutex) {

			mu.Lock()
			md5 := DataSignerMd5(inString)
			mu.Unlock()

			out_ch1 := make(chan string, 1)
			out_ch2 := make(chan string, 1)

			DataSignerCrc32Routine := func(inString string, out chan string) {
				out <- DataSignerCrc32(inString)
			}

			go DataSignerCrc32Routine(inString, out_ch1)
			go DataSignerCrc32Routine(md5, out_ch2)

			str1 := <-out_ch1
			str2 := <-out_ch2

			retStr := str1 + "~" + str2

			out <- retStr

			atomic.AddUint32(&recieved, ^uint32(0))
			if atomic.LoadUint32(&recieved) == 0 {
				//close(out)
			}

		}(strconv.Itoa(i.(int)), mu)
	}

	//fmt.Println("\t SingleHash valueCounter", valueCounter)

})

var MultiHash = job(func(in, out chan interface{}) {

	var recieved uint32

	for dataIm := range in {
		atomic.AddUint32(&recieved, 1)
		//recieved++
		go func(dataIm interface{}, out chan interface{}) {

			fmt.Println("\tMultiHash get", dataIm)

			wg := &sync.WaitGroup{}
			mu := &sync.Mutex{}
			//resChan := make(chan string, 6)

			var resArray [6]string

			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func(inString string, th int, wg *sync.WaitGroup, resArray *[6]string) {
					defer wg.Done()
					strTemp := DataSignerCrc32(strconv.Itoa(th) + inString)

					mu.Lock()
					resArray[th] = strTemp
					mu.Unlock()
					//runtime.Gosched()
				}(dataIm.(string), i, wg, &resArray)
			}

			wg.Wait()

			//runtime.Gosched()

			retStr := strings.Join(resArray[:], "")
			fmt.Println("\tMultiHash out", retStr)
			out <- retStr

			atomic.AddUint32(&recieved, ^uint32(0))
			if atomic.LoadUint32(&recieved) == 0 {
				//close(out)
			}

		}(dataIm, out)

	}

	//fmt.Println("\t MultiHash valueCounter", valueCounter)
	//in <- valueCounter

})

//* CombineResults получает все результаты, сортирует (https://golang.org/pkg/sort/),
//объединяет отсортированный результат через _ (символ подчеркивания) в одну строку

var CombineResults = job(func(in, out chan interface{}) {
	var combineSlice []string
	for dataIm := range in {
		combineSlice = append(combineSlice, dataIm.(string))
	}

	sort.Strings(combineSlice)

	retStr := strings.Join(combineSlice, "_")
	fmt.Println("\t CombineResults out", retStr)

	out <- retStr
})

/*
var recieved uint32
 var freeFlowJobs = []job{
	job(func(in, out chan interface{}) {
		out <- uint32(1)
		out <- uint32(3)
		out <- uint32(4)
	}),
	job(func(in, out chan interface{}) {
		for val := range in {
			out <- val.(uint32) * 3
			time.Sleep(time.Millisecond * 100)
		}
	}),
	job(func(in, out chan interface{}) {
		for val := range in {
			fmt.Println("collected", val)
			atomic.AddUint32(&recieved, val.(uint32))
		}
	}),
}
*/

func ExecutePipeline(in ...job) (result int) {
	//fmt.Printf("in := %#v \n", in)

	in_ch := make(chan interface{}, 100)
	out_ch := make(chan interface{}, 100)

	//var wg sync.WaitGroup
	//const numDigesters = 20
	//wg.Add(1)

	//go func() {
	in[0](in_ch, out_ch)
	in_ch = out_ch
	//	wg.Done()
	//}()

	//close(out_ch)

	//go func() {
	//	wg.Wait()
	close(out_ch)
	//}()

	out_ch = make(chan interface{}, 100)

	count := CountValuesFromChannel(in_ch, out_ch)
	in_ch = out_ch
	out_ch = make(chan interface{}, 100)

	for i := 1; i < len(in); i++ {
		ReadNValuesAndCloseChannel(in_ch, out_ch, count)
		in_ch = out_ch
		out_ch = make(chan interface{}, 100)
		in[i](in_ch, out_ch)
		in_ch = out_ch
		out_ch = make(chan interface{}, 100)

	}

	return
}

func CountValuesFromChannel(in, out chan interface{}) int {

	var count int

	for i := range in {
		count++
		out <- i.(interface{})
	}
	close(out)
	return count
}

func ReadNValuesAndCloseChannel(in, out chan interface{}, n int) {

	var wg sync.WaitGroup
	//const numDigesters = 20
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			t := <-in
			out <- t.(interface{})
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
}

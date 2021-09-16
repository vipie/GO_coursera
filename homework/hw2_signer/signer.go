package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func main() {

}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, singleJob := range jobs {
		wg.Add(1)
		out := make(chan interface{})
		go jobMaker(singleJob, in, out, wg)
		in = out
	}
	wg.Wait()
}

func jobMaker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)
	job(in, out)
}

func crc32signer(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for v := range in {
		wg.Add(1)
		go singleHashJob(v.(int), out, wg, mu)
	}
	wg.Wait()
}

func singleHashJob(value int, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	data := strconv.Itoa(value)

	rrr := make(chan string)

	go func(data string, out chan string) {

		mu.Lock()
		md5 := DataSignerMd5(data)
		mu.Unlock()
		out <- md5
	}(data, rrr)

	crc32Chan := make(chan string)
	go crc32signer(data, crc32Chan)

	crc32md5Chan := make(chan string)

	md5r := <-rrr

	go func(out chan string, data string) {
		out <- DataSignerCrc32(data)
	}(crc32md5Chan, md5r)

	crc32md5 := <-crc32md5Chan
	crc32 := <-crc32Chan

	//result := crc32 + "~" + crc32md5

	/*fmt.Printf("%s SingleHash data %s\n", data, data)
	fmt.Printf("%s SingleHash md5(data) %s\n", data, md5)
	fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32md5)
	fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32)
	fmt.Printf("%s SingleHash result %s\n\n", data, result)*/
	out <- crc32 + "~" + crc32md5
}

func MultiHash(in, out chan interface{}) {
	maxTh := 6
	wg := &sync.WaitGroup{}
	for v := range in {
		wg.Add(1)
		go multiHashJob(v.(string), out, wg, maxTh)
	}
	wg.Wait()
}

func multiHashJob(value string, out chan interface{}, wg *sync.WaitGroup, maxTh int) {
	defer wg.Done()
	results := make([]string, maxTh)
	wgJob := &sync.WaitGroup{}
	muJob := &sync.Mutex{}
	for th := 0; th < maxTh; th++ {
		wgJob.Add(1)
		data := strconv.Itoa(th) + value
		go func(data string, th int, results []string, wgJob *sync.WaitGroup, muJob *sync.Mutex) {
			result := DataSignerCrc32(data)
			muJob.Lock()
			results[th] = result
			muJob.Unlock()
			//fmt.Printf("%s MultiHash: crc32(th+step1)) %v %s\n", value, th, results[th])
			wgJob.Done()
		}(data, th, results, wgJob, muJob)
	}
	wgJob.Wait()
	//resultStr :=
	//fmt.Printf("%s MultiHash result: %s\n\n", value, resultStr)
	out <- strings.Join(results[:], "")
}

func CombineResults(in, out chan interface{}) {
	resultsSlice := make([]string, 0)
	for v := range in {
		resultsSlice = append(resultsSlice, v.(string))
	}
	sort.Strings(resultsSlice)
	//result :=
	//fmt.Printf("CombineResults %s\n", result)
	out <- strings.Join(resultsSlice[:], "_")
}

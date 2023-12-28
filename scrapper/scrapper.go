package scrapper

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)
type extractedJob struct{
	title string
	url string
	location string
	company string
	summary string
}

// Scrape JobsDB by a term
func Scrape(term string){
	var baseURL string = "https://hk.jobsdb.com/hk/search-jobs/" + term
	var jobs[] extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	
	for i :=0; i <totalPages; i++{
		go getPage(i,baseURL,c)
	}

	for i:=0; i <totalPages;i++{
		extractedJobs:= <-c
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
}

func getPage(page int, url string, mainC chan<- []extractedJob){
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageUrl := url + "/" + strconv.Itoa(page)
	res,err := http.Get(pageUrl)

	checkErr(err)
	checkCode(res)
	
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	// URL, title, company
	searchCards := doc.Find(".z1s6m00 ._1hbhsw6ce")
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, c)
		// location := s.Find("div + span a").First().Text()
		// fmt.Println(location)
	})

	for i :=0; i < searchCards.Length(); i++{
		job := <-c
		jobs = append(jobs,job)
	}
	mainC <- jobs

	//Location	
	// location := doc.Find(".z1s6m00 ._1hbhsw64y .y44q7i0 .y44q7i3 .y44q7i21 .y44q7ih")
	// fmt.Println("location1 : " , location)
	// location.Each(func(i int, k *goquery.Selection){
	// 	locations := k.Find("a").Text()
	// 	fmt.Println("location : " + locations)
	// 	})
	
	//Summary
	// outDiv := doc.Find(".z1s6m00 ._1hbhsw6ba ._1hbhsw64y")
	// ul := outDiv.Find(".z1s6m00 .z1s6m03 ._5135ge0 ._5135ge5")
	// li := ul.Find(".z1s6m00 ._1hbhsw66q").First()
	// div := li.Find(".z1s6m00 ._1hbhsw6r ._1hbhsw6p ._1hbhsw6a2")
	// div.Each(func(i int, summ *goquery.Selection){
	// 	summaryInfo := summ.Find("span").Text()
	// 	fmt.Println(summaryInfo)
	// })
		
}

func extractJob(card *goquery.Selection, c chan<- extractedJob){
	anchor := card.Find("a")
	url, _ := anchor.Attr("href")
	title := anchor.Find("span").First().Text()
	company := card.Find("span>a").First().Text()
	location :="empty location"
	summary := "empty summary"

	c <- extractedJob{
		url : url,
		title: title,
		company : company,
		location: location,
		summary: summary,
	}
}

func getPages(url string) int {
	pages :=0
	res,err := http.Get(url)
	
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find("._1hbhsw66m").Each(func(i int, page *goquery.Selection){
		if page.Length() > 0{
			pages = page.Find("option").Length()
			}
	})
	return pages
}

func writeJobs(jobs []extractedJob){
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"TITLE", "URL", "COMPANY", "LOCATION", "SUMMARY"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs{
		jobSlice := []string{job.title,"https://hk.jobsdb.com" +job.url, job.company, job.location, job.summary}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkErr(err error){
	if err != nil{
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response){
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status: " , res.StatusCode)
	}
}

// CleanString cleans a string
func CleanString(str string) string{
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
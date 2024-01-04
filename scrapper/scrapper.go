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
	summary[] string
}

// Scrape JobsDB by a term
func Scrape(term string){
	var baseURL string = "https://hk.jobsdb.com/hk/search-jobs/" + term
	var jobs[] extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	
	for i :=1; i <= totalPages; i++{
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
	searchCards := doc.Find(".z1s6m00 ._1hbhsw6ce > .z1s6m00[data-automation='jobListing'] > div")
	searchCards.Each(func(i int, card *goquery.Selection){
		go extractJob(card, c)
	})

	for i :=0; i < searchCards.Length(); i++{
		job := <-c
		jobs = append(jobs,job)
	}
	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob){
	var (
		url,title string
	) 
	var summary[] string
		
	// url, title
	infoA := card.Find(".z1s6m00 ._1hbhsw6ce > h1")
	infoA.Each(func(i int, info *goquery.Selection){
		if info.Length() > 0{
			anchor := info.Find("a")
			url, _ = anchor.Attr("href")
			title = anchor.Find("span").First().Text()
		}
	})
	
	// company, location, summary
	company := card.Find("span>a").First().Text()
	locationInfo := card.Find("span.z1s6m00._1hbhsw64y.y44q7i0.y44q7i3.y44q7i21.y44q7ih")
	location := locationInfo.Find("a").Text()

	summaryDiv := card.Find("ul > li")
	summaryDiv.Each(func(i int, li *goquery.Selection){
		summaryInfo :=li.Find("span").Text()
		summary = append(summary, summaryInfo)
	})

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

	doc.Find(".z1s6m00 ._1hbhsw6ce ._1hbhsw6p").Each(func(i int, page *goquery.Selection){
			pageStr := page.Find("option").Last().Text()
			if pageValue, err := strconv.Atoi(pageStr); err == nil && pageValue > 0{
				checkErr(err)
				pages += pageValue
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
		jobSlice := []string{job.title,"https://hk.jobsdb.com" +job.url, job.company, job.location, strings.Join(job.summary, "\n")}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func checkErr(err error){
	if err != nil{
		log.Fatalln("err : ", err)
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
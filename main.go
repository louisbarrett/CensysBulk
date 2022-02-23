package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"louisbarrett/censysbulk/querylist"
	"net/http"
	"os"
	"strings"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
)

var (
	UserAgentString = "recontools/1.0"
	CensysAPIKey    = os.Getenv("CENSYSAPIKEY")
	CensysSecret    = os.Getenv("CENSYSAPISECRET")
	CensysQuery     = flag.String("query", "services.kubernetes.endpoints.name:*", "Censys query")
	flagQueryList   = flag.Bool("l", false, "List all available Censys fields")
	flagVerbose     = flag.Bool("v", false, "Print verbose errors")
)

func init() {
	// check API key
	if CensysAPIKey == "" || CensysSecret == "" {
		log.Fatal("Censys API key not set, set it using the CENSYSAPIKEY and CENSYSAPISECRET environment variables")
	}

}

func main() {
	// remove problematic characters from query
	*CensysQuery = strings.Replace(*CensysQuery, "`", "'", -1)

	flag.Parse()
	if *flagQueryList {
		fmt.Println(querylist.QueryDefinitions)
		return
	}

	if *flagVerbose {
		log.Println(*CensysQuery)
	}
	queryResults := bulkQueryCensys(*CensysQuery)
	fmt.Println(queryResults.StringIndent("", "  "))

}

func bulkQueryCensys(query string) (allResults gabs.Container) {
	var results gabs.Container
	results, nextPage := queryCensys(query, "")
	allResults.ArrayAppend(results.Data())
	var counter = 0
	for nextPage != "" {
		results, nextPage = queryCensys(query, nextPage)
		allResults.ArrayAppend(results.Data())
		counter++
	}

	return allResults
}

func queryCensys(query string, cursor string) (results gabs.Container, nextCursor string) {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}
	httpRequestBody := `
	{
		"q": "%s",
		"per_page": 100,
		"cursor": "%s",
		"fields": [
			"location.registered_country",
			"location.longitude",
			"location.continent",
			"url",
			"ip",
			"location.registered_country_code",
			"location.country_code",
			"location.latitude",
			"protocols"
		]
	}`

	httpRequestBody = fmt.Sprintf(httpRequestBody, query, cursor)
	httpRequestData, err := gabs.ParseJSON([]byte(httpRequestBody))
	if *flagVerbose {
		if err != nil {
			log.Fatal(err)
		}
	}
	requestBodyBytes := httpRequestData.Bytes()
	requestBodyReader := bytes.NewReader(requestBodyBytes)

	httpRequest, err := http.NewRequest("POST", "https://search.censys.io/api/v2/hosts/search", requestBodyReader)
	if err != nil {
		log.Fatal(err)
	}
	httpRequest.SetBasicAuth(CensysAPIKey, CensysSecret)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Add("User-Agent", UserAgentString)

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		log.Fatal("Error ", err)
	}

	responseBytes := httpResponse.Body
	message, err := ioutil.ReadAll(responseBytes)
	prettyPrint, err := gabs.ParseJSON(message)
	if err != nil {
		log.Fatal("Error ", string(message), err)
	}
	if *flagVerbose {
		// Print results of query to STDOUT
		fmt.Println(string(prettyPrint.Path("result.hits").String()))
	}
	if prettyPrint.Path("result.links.next").Data() != nil {
		nextCursor = prettyPrint.Path("result.links.next").Data().(string)
	} else {
		log.Fatal("Error ", string(message), err)
	}
	return *prettyPrint.Path("result.hits"), nextCursor
}

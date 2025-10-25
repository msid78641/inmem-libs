package to_do

import (
	"fmt"
	"io"
	"net/http"
	url "net/url"
)

func GetToDoLoader(key string) (interface{}, error) {
	baseUrl, _ := url.Parse("https://jsonplaceholder.typicode.com/todos/")
	baseUrl = baseUrl.JoinPath(key)
	resp, err := http.Get(baseUrl.String())
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return nil, err
	}

	// IMPORTANT: Defer closing the response body to prevent resource leaks
	defer resp.Body.Close()

	// 3. Check the Status Code
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Received non-OK status code: %d\n", resp.StatusCode)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}
	return string(body), nil
}

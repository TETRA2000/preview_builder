package preview_builder

import "github.com/google/go-github/github"

func main()  {
	client := github.NewClient(nil)

	// list all organizations for user "willnorris"
	_, _, _ = client.Organizations.List("willnorris", nil)
}
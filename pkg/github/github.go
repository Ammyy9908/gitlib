package github

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Ammyy9908/gitlib/pkg/models"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type GithubService struct {
	client  *github.Client
	owner   string
	repo    string
	context context.Context
}

func NewGithubService(token, owner, repo string) *GithubService {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := github.NewClient(oauthClient)
	return &GithubService{
		client:  client,
		owner:   owner,
		repo:    repo,
		context: context.Background(),
	}

}

func (g *GithubService) AddCollaborator(username string) error {
	_, _, err := g.client.Repositories.AddCollaborator(context.Background(), g.owner, g.repo, username, nil)
	return err
}

func (g *GithubService) ViewUserProfile(username string) (*models.Profile, error) {
	user, _, err := g.client.Users.Get(context.Background(), username)
	return &models.Profile{Name: user.GetName(), Email: user.GetEmail()}, err
}

func (g *GithubService) ShareCode(username, featureName, codeContent string, AccessToken string) error {
	branchName := username
	baseUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s", g.owner, g.repo)

	// 1. Get reference to main branch
	resp, err := http.Get(fmt.Sprintf("%s/git/refs/heads/main", baseUrl))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var refData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&refData); err != nil {
		return err
	}

	mainSHA := refData["object"].(map[string]interface{})["sha"].(string)

	// 2. Create a new branch
	branchCreationData := map[string]interface{}{
		"ref": fmt.Sprintf("refs/heads/%s", branchName),
		"sha": mainSHA,
	}

	jsonData, _ := json.Marshal(branchCreationData)
	fmt.Println(string(jsonData))
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/git/refs", baseUrl), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken)) // Assuming g.token is the authentication token
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	//print the response body
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	// 3. Create a new blob (representing the file)
	fileContent := []byte(codeContent)
	blobCreationData := map[string]string{"content": base64.StdEncoding.EncodeToString(fileContent), "encoding": "base64"}

	jsonData, _ = json.Marshal(blobCreationData)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/git/blobs", baseUrl), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var blobData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&blobData); err != nil {
		return err
	}

	blobSHA := blobData["sha"].(string)

	// 4. Create a new tree with the blob
	treeData := map[string][]map[string]string{
		"tree": {{
			"path": "filename.txt", // Your file name
			"mode": "100644",       // Blob (file) mode
			"type": "blob",
			"sha":  blobSHA,
		}},
	}

	jsonData, _ = json.Marshal(treeData)
	fmt.Println("Tree Request", string(jsonData))
	fmt.Print("Base url", fmt.Sprintf("%s/git/trees", baseUrl))
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/git/trees", baseUrl), bytes.NewBuffer(jsonData))
	// ... (similar to before, get the tree SHA)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var treeResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&treeResponse); err != nil {
		return err
	}

	fmt.Println("Tree response", treeResponse)

	treeSHA := treeResponse["sha"].(string)

	// 5. Create a new commit
	commitData := map[string]interface{}{
		"message": "Your commit message here",
		"tree":    treeSHA,
		"parents": []string{mainSHA},

		// ... (other commit details, including parent SHA and tree SHA)
	}

	jsonData, _ = json.Marshal(commitData)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/git/commits", baseUrl), bytes.NewBuffer(jsonData))
	// ... (similar to before, get the commit SHA)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var commitResponse map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&commitResponse); err != nil {
		return err
	}
	//use the request

	fmt.Println("Commit response", commitResponse)

	commitSha := commitResponse["sha"].(string)
	// 6. Update the reference to point to the new commit
	refUpdateData := map[string]string{"sha": commitSha}

	jsonData, _ = json.Marshal(refUpdateData)
	req, _ = http.NewRequest("PATCH", fmt.Sprintf("%s/git/refs/heads/%s", baseUrl, branchName), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// 7. Create a Pull Request
	prData := map[string]interface{}{
		"title": "New Code Submission from " + username,
		"body":  "Description for the PR",
		"head":  branchName,
		"base":  "main", // or "master" or whichever branch you're targeting
	}

	jsonData, _ = json.Marshal(prData)
	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/pulls", baseUrl), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", AccessToken))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var prResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&prResponse); err != nil {
		return err
	}

	// You can capture the PR URL or other relevant details from the response if needed.
	fmt.Print("Pull Request Created Successfully")

	return nil
}

package preview_builder

import (
	"github.com/google/go-github/github"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/oauth2"
	"fmt"
	"os/exec"
	"os"
	"time"
	"errors"
	"regexp"
	"database/sql"
)

func CreateGithubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ""},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(tc)
}

func GetPRList(client **github.Client) ([]*github.PullRequest, error) {
	opt := &github.PullRequestListOptions{}
	opt.State = "open"
	opt.PerPage = 100
	opt.Sort = "updated"
	opt.Direction = "desc"

	list, _, err := (*client).PullRequests.List("TETRA2000", "reponame", opt)
	return list, err
}

func GetListCommits(client **github.Client, prNumber int) ([]*github.RepositoryCommit, error) {
	commitOpt := &github.ListOptions{Page:1, PerPage:100}
	commits, _, err := (*client).PullRequests.ListCommits("TETRA2000", "reponame", prNumber, commitOpt)
	return commits, err
}

func OpenSqliteDb() *sql.DB {
	db, err := sql.Open("sqlite3", "./preview_builder_data.db")
	if err != nil {
		fmt.Print(err)
	}

	return db
}

func ImageExists(imageName string) bool {
	out, err := exec.Command("docker", "images", "-q", imageName).Output()

	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	return  string(out) != ""
}

func ImageCreatedAt(imageName string) (time.Time, error) {
	if !ImageExists(imageName) {
		return time.Now(), errors.New("The image doesn't exist.")
	}

	rawout, err := exec.Command("docker", "inspect", "-f", "{{ .Created }}", imageName).Output()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	re := regexp.MustCompile(`\r?\n`)
	out := re.ReplaceAllString(string(rawout), "")


	layout := "2006-01-02T15:04:05.999999999Z"
	return time.Parse(layout, out)
}

func BuildPreviewImage(number int) string {
	// Same as build_pr_image.sh
	RunCommandInDir("web", "git", "fetch")
	RunCommandInDir("web", "git", "checkout", "origin/master")
	RunCommandInDir("web", "git", "pull", "--no-edit", "-s", "ours", "origin", fmt.Sprintf("pull/%d/head", number))
	RunCommandInDir("web", "git", "gc")
	//FIXME linux only (timeout command).
	out, err := exec.Command("timeout", "15m", "docker", "build", "-t", ImageName(number), ".").Output()
	if err != nil {
		fmt.Print(err)
		fmt.Print(string(out))
	}
	return string(out)
}

func StartPreviewContainer(number int) string {
	if !ImageExists(ImageName(number)) {
		fmt.Print("Image doesn't exist!!")
		os.Exit(1)
	}

	// ignore errors
	exec.Command("docker", "stop", ContainerName(number)).Run()
	exec.Command("docker", "rm", ContainerName(number)).Run()

	cmd := exec.Command("docker", "run", "--name", ContainerName(number), "-d", "--restart=always", "-p", fmt.Sprintf("%d:80", 10000 +number), "-p", fmt.Sprintf("%d:8080", 30000 +number), ImageName(number))
	out, err := cmd.Output()
	if err != nil {
		fmt.Print(err)
		fmt.Print(string(out))
		os.Exit(1)
	}

	return string(out)
}

func RunCommand(name string, arg ...string) {
	out, err := exec.Command(name, arg...).Output()

	if err != nil {
		fmt.Print(err)
		fmt.Print(string(out))
		os.Exit(1)
	}
}

func RunCommandInDir(dir string, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir

	out, err := cmd.Output()

	if err != nil {
		fmt.Print(err)
		fmt.Print(string(out))
		os.Exit(1)
	}
}

func ImageName(number int) string {
	return fmt.Sprintf("preview_builder:%d", number)
}

func ContainerName(number int) string {
	return fmt.Sprintf("preview_builder_pr_%d", number)
}

func StoreUpdatedAt(db  *sql.DB, pr *github.PullRequest) {
	_, err := db.Exec(
		"INSERT or REPLACE into pull_request (number,latest_commit_sha1 , updated_at) VALUES (?, (SELECT latest_commit_sha1 FROM pull_request WHERE number = ?), ?)",
		*pr.Number,
		*pr.Number,
		(*pr.UpdatedAt).Format(time.RFC3339Nano),
	)
	if err != nil {
		fmt.Println(err)
	}
}

//func addComment(client **github.Client, number int, comment **github.IssueComment)  {
//	//_,_, err := (*client).Issues.CreateComment("TETRA2000", "reponame", number, *comment)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
//
//func addLabel(client **github.Client, number int, label string)  {
//	//issue, _, err := (*client).Issues.Get("TETRA2000", "reponame", number)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	labels := []string{label}
//	for _, issueLabel := range issue.Labels {
//		labels = append(labels, *issueLabel.Name)
//	}
//
//
//	issueRequest := &github.IssueRequest{Labels: &labels}
//
//	//_,_, err = (*client).Issues.Edit("TETRA2000", "reponame", number, issueRequest)
//	if err != nil {
//		fmt.Println(err)
//	}
//}
//
//func updateCommitStatus(client **github.Client, commit_hash string, status string)  {
//
//}
//
//func createReview(client **github.Client, number int, review **github.PullRequestReviewRequest)  {
//    //_,_, err :=client.Issues.CreateComment("TETRA2000", "reponame", number, comment)
//    if err != nil {
//        fmt.Println(err)
//    }
//}
//
//func formatBuildLog(gulpOutput string) string {
//	return fmt.Sprintf("<details><summary>gulp log (click to expand)</summary><pre>%s</pre></details>", gulpOutput)
//}
//
//
//func parseGulpError(gulpOutput string) string  {
//	gulpErrorRP := regexp.MustCompile(`(?i)error`)
//	gulpErrorLineRP := regexp.MustCompile(`Error`)
//	gulpEmptyLineRP := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\]\s{2}`)
//	gulpHTMLLintLineRP := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\]\s\d+\serror.*`)
//	gulpVarionsErrorLineRP := regexp.MustCompile(`【ERROR】.*`)
//	gulpFinishLineRP := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\].*Finished.*build.*`)
//
//	var errorBody string
//	if (strings.Index(gulpOutput, "✖") != -1 || gulpErrorRP.MatchString(gulpOutput)) {
//
//		gulpErrorBeginLine := -1
//		if gulpEmptyLineRP.MatchString(gulpOutput) {
//			gulpErrorBeginLine  = gulpEmptyLineRP.FindStringSubmatchIndex(gulpOutput)[0]
//		}
//
//		if gulpHTMLLintLineRP.MatchString(gulpOutput) {
//			lintErrorStartLine  := gulpHTMLLintLineRP.FindStringSubmatchIndex(gulpOutput)[0]
//			if gulpErrorBeginLine == -1 || lintErrorStartLine < gulpErrorBeginLine {
//				gulpErrorBeginLine = lintErrorStartLine
//			}
//		}
//
//		if gulpVarionsErrorLineRP.MatchString(gulpOutput) {
//			variousErrorStartLine  := gulpVarionsErrorLineRP.FindStringSubmatchIndex(gulpOutput)[0]
//			if gulpErrorBeginLine == -1 || variousErrorStartLine < gulpErrorBeginLine {
//				gulpErrorBeginLine = variousErrorStartLine
//			}
//		}
//
//		if gulpErrorBeginLine != -1 {
//			gulpErrorLineIndexes := gulpErrorLineRP.FindStringSubmatchIndex(gulpOutput)
//			if len(gulpErrorLineIndexes) > 0 && gulpErrorBeginLine > gulpErrorLineIndexes[0] {
//				gulpErrorBeginLine = gulpErrorLineIndexes[0]
//			}
//
//			for _, positions := range gulpFinishLineRP.FindAllStringSubmatchIndex(gulpOutput, -1) {
//				if (positions[0] > gulpErrorBeginLine) {
//					errorBody = gulpOutput[gulpErrorBeginLine:(positions[0])]
//					break
//				}
//			}
//		} else {
//			// Failed to detect start line of errors
//			errorBody = "There is '✖' or 'error' in your build result."
//		}
//	}
//	return errorBody
//}
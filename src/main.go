package main

import (
	"preview_builder"
	"fmt"
	"os"
	"time"
)

func main() {
	db := preview_builder.OpenSqliteDb()
	defer db.Close()

	client := preview_builder.CreateGithubClient()

	prlist, err := preview_builder.GetPRList(&client)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	for _, pr := range prlist {
		var updated_at_local string
		rows, err := db.Query("select updated_at from pull_request where number = ?", *pr.Number)
		if err == nil && rows.Next() {
			_ = rows.Scan(&updated_at_local)
		}
		rows.Close()

		if updated_at_local != "" {
			t, err := time.Parse(time.RFC3339Nano, updated_at_local)
			if err == nil && t.Equal(*pr.UpdatedAt) {
				continue
			}
		}

		fmt.Print(*pr.Number)
		fmt.Printf(" : ")

		prDetail, _, err := client.PullRequests.Get("TETRA2000", "reponame", *pr.Number)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}
		if prDetail.Mergeable == nil || *prDetail.Mergeable == false {
			fmt.Println("Skipped. This is unmergeable.")
			// マージ判定に時間がかかる場合があるので、ここではPR更新時刻を保存しない
			continue
		}

		commits, err := preview_builder.GetListCommits(&client, *pr.Number)
		if err != nil {
			fmt.Print(err)
			os.Exit(-1)
		}

		if (len(commits) > 0) {
			rows, err := db.Query("select latest_commit_sha1 from pull_request where number = ?", *pr.Number)
			if err != nil {
				fmt.Println(err)
			}

			var sha1_at_last_build string
			if rows.Next() {
				err = rows.Scan(&sha1_at_last_build)
				if err != nil {
					fmt.Println(err)
				}
			}
			rows.Close()

			latest_commit := *(commits[len(commits)-1])
			if sha1_at_last_build == *latest_commit.SHA {
				fmt.Println("Skipped. Commit didn't change.")
				preview_builder.StoreUpdatedAt(db, pr)
				continue
			}

			buildOutput := preview_builder.BuildPreviewImage(*pr.Number)
			fmt.Print("Build completed!!")
			fmt.Println(buildOutput)

			_, err = db.Exec(
				"INSERT or REPLACE into pull_request (number, latest_commit_sha1, updated_at) VALUES (?, ?, ?)",
				*pr.Number,
				*latest_commit.SHA,
				(*pr.UpdatedAt).Format(time.RFC3339Nano),
			)
			if err != nil {
				fmt.Println(err)
			}
		}

		fmt.Printf("\n")
	}
}


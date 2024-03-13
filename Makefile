-include .env
export

run:
	@go run main.go -command worklog -i

bulk-run:
	@go run main.go -command worklog

my-issues:
	@go run main.go -command my-issues
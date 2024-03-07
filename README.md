# Setup

Setup environment variables in .env file

```
JIRA_AUTH_TOKEN=
MEETINGS_JIRA_ISSUE=
JIRA_HOST=
```

1. Set your jira host, example
```example
  JIRA_HOST=jira.com
```
2. Set your personal auth token. You can create one in https://{JIRA_HOST}/secure/ViewProfile.jspa
3. (Optional) Set default jira issue. It will be used if the task is empty. It can be used to log time for meetings, code review, etc

# Run

## Interactive mode
Will sequentially ask you to input issue, date, spent time and comment.

```shell
  make run
```

Input examples:
1. All params set
```
Issue: INT-1
Date: 2024-02-12
Spent time: 1h30m
Comment: Test worklog
```
2. Use default values, default issue, current date, 1 hour and no comment
```
Issue: 
Date: 
Spent time: 1
Comment:
```

## Bulk mode
Will ask you to input worklog as row {YYYY-MM-DD} {ISSUE} {SPENT_TIME} {COMMENT}

```shell
  make bulk-run
```

## Send

After input program show you input result and ask you if it's ok to send result. Press "n" to not send or enter to send result.
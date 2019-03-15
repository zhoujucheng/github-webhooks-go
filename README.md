# github-webhooks-go
Web server for github webhooks.

## Usage
 ```
  -c string
        config path (default "/etc/github-webhooks/config")
  -h    show this help
  -port int
        local port to listen. (default 9966)
```

## Config
Each line is a webhook. The first column is URL relative path, second column is secrect key, third column is executable file path.
For example:/github/webhooks aabbcc123 /usr/local/bin/updatefromrepo, 
The following code will be executed when receive a webhook.
```
/usr/local/bin/updatefromrepo <eventtype> <requestbody>
```
The \<eventtype\> come from request header: X-GitHub-Event. See https://developer.github.com/webhooks/#events

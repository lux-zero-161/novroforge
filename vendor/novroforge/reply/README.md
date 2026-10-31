[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/code.forgejo.org/forgejo/reply)
[![Go Report Card](https://goreportcard.com/report/code.forgejo.org/forgejo/reply)](https://goreportcard.com/report/code.forgejo.org/forgejo/reply)

# reply

 Library to trim replies from plain text email. (Golang port of https://github.com/discourse/email_reply_trimmer)

 Forked from https://github.com/dimiro1/reply

 ## Why

 Reply was forked because the original project does not seem to be maintained.
 Furthermore, the version before the fork removes parts of the mail that should not be removed.
 The project was forked to change this.

# Usage

```go
package main

import (
    "fmt"

    "code.forgejo.org/forgejo/reply"
)

func main() {
    message := `
        This is before the embedded email.
        
        On Wed, Sep 25, 2013, at 03:57 PM, richard_clark wrote:
        
        Richard> This is the embedded email
        
        This is after the embedded email and will not show up because 99% of the times
        this is the signature...
    `
	fmt.Println(reply.FromText(message))
}
```

will output:

```text
This is before the embedded email.
```

# pasw - Parse Swagger

pasw is a simple tool that parses swagger json structure and outputs ready to use curl/ffuf commands.

## Install

For Go 1.17+

```
go install -v git.sr.ht/~ohdude/pasw/cmd/pasw@latest
```

Alternatively:

```
git clone https://git.sr.ht/~ohdude/pasw
cd pasw
make
```

## Usage

```
$ cat test.json | pasw
curl -X GET https://test.com/v1/company/profiles/{id}
curl -X DELETE https://test.com/v1/company/profiles/{id}

$ cat test.json | pasw -o ffuf
ffuf -X GET -u https://test.com/v1/company/profiles/{id}
ffuf -X DELETE -u https://test.com/v1/company/profiles/{id}
```
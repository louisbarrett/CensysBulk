# Censys Bulk Search

Extracts multi page results from https://search.censys.io using the v2 API. 

## Install

Set your API key via environment variables

```
export CENSYSAPIKEY=<your api key>
export CENSYSAPISECRET=<your api secret>
```

Pull the source code using go get
```
go get github.com/louisbarrett/censysbulk
```

Install using go install
```
go install github.com/louisbarrett/censysbulk
```

## Usage

```
Usage of ./censysbulk:
  -l    List all available Censys fields
  -query string
        Censys query (default "services.kubernetes.endpoints.name:*")
  -v    Print verbose errors
```

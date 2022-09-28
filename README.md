# email-recruitment

A simple go program that send recruitment emails based on a Go `html/template` template.

Usage:

```bash
cd data_example
go run ../cmd/assign -reassign=true # or you can go to prospects.json and manually assign recruiters
go run ../cmd/mail -diverge=output # remove this flag to actually send emails out
```
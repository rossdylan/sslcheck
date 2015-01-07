# sslcheck

[![Build Status](https://travis-ci.org/rossdylan/sslcheck.png)](https://travis-ci.org/rossdylan/sslcheck)

# Features
- Super Simple
- Check a list of services to see if their certs have expired
- Print a pretty table with the relevant information
- Email out a report to a given address

#Example:

```
    $ sslcheck csh.rit.edu:443 google.com:443
    $ sslcheck -file sites.txt
```

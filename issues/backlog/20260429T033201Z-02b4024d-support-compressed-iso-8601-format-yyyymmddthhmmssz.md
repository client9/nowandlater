{
  "title": "support compressed ISO 8601 format YYYYMMDDThhmmssZ 20260429T030444Z",
  "id": "20260429T033201Z-02b4024d",
  "state": "backlog",
  "created": "2026-04-29T03:32:01Z",
  "labels": [],
  "assignees": [],
  "milestone": "",
  "projects": [],
  "template": "",
  "events": [
    {
      "ts": "2026-04-29T03:32:01Z",
      "type": "filed",
      "to": "backlog"
    }
  ]
}

currently supported is standard ISO 8601 format '2026-04-29T03:04:44Z'

```sh
$ go run cmd/nldate/main.go '2026-04-29T03:04:44Z'
input:     "2026-04-29T03:04:44Z"
signature: "YEAR INTEGER INTEGER TIME TIMEZONE"
tokens:
  [0] YEAR           2026
  [1] INTEGER        4
  [2] INTEGER        29
  [3] TIME           "03:04:44"
  [4] TIMEZONE       "z"
period:    second
now:       2026-04-28T20:32:55-07:00
resolved:  2026-04-29T03:04:44Z
```

However the compressed format is not '20260429T030444Z'  (YYYYMMDDThhmmssZ).  It's a fixed format with 'T' in the middle and 'Z' at the end, total 18 characters.

```sh
$ go run cmd/nldate/main.go '20260429T030444Z'
input:     "20260429T030444Z"
signature: "INTEGER UNKNOWN"
tokens:
  [0] INTEGER        20260429
  [1] UNKNOWN        "030444z"
parse:     error: nowandlater: unknown date signature
```

This is special case fixed format date and time representation.  There is no internationalization or special langauge support needed.



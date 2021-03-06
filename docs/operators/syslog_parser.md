## `syslog_parser` operator

The `syslog_parser` operator parses the string-type field selected by `parse_from` as syslog. Timestamp parsing is handled automatically by this operator.

### Configuration Fields

| Field         | Default          | Description                                                                                                                                |
| ---           | ---              | ---                                                                                                                                        |
| `id`          | `syslog_parser`  | A unique identifier for the operator                                                                                                       |
| `output`      | Next in pipeline | The connected operator(s) that will receive all outbound entries                                                                           |
| `parse_from`  | $                | A [field](/docs/types/field.md) that indicates the field to be parsed as JSON                                                              |
| `parse_to`    | $                | A [field](/docs/types/field.md) that indicates the field to be parsed as JSON                                                              |
| `preserve_to` |                  | Preserves the unparsed value at the specified [field](/docs/types/field.md)                                                                |
| `on_error`    | `send`           | The behavior of the operator if it encounters an error. See [on_error](/docs/types/on_error.md)                                            |
| `protocol`    | required         | The protocol to parse the syslog messages as. Options are `rfc3164` and `rfc5424`                                                          |
| `timestamp`   | `nil`            | An optional [timestamp](/docs/types/timestamp.md) block which will parse a timestamp field before passing the entry to the output operator |
| `severity`    | `nil`            | An optional [severity](/docs/types/severity.md) block which will parse a severity field before passing the entry to the output operator    |

### Example Configurations


#### Parse the field `message` as syslog

Configuration:
```yaml
- type: syslog_parser
  protocol: rfc3164
```

<table>
<tr><td> Input record </td> <td> Output record </td></tr>
<tr>
<td>

```json
{
  "timestamp": "",
  "record": "<34>Jan 12 06:30:00 1.2.3.4 apache_server: test message"
}
```

</td>
<td>

```json
{
  "timestamp": "2020-01-12T06:30:00Z",
  "record": {
    "appname": "apache_server",
    "facility": 4,
    "hostname": "1.2.3.4",
    "message": "test message",
    "msg_id": null,
    "priority": 34,
    "proc_id": null,
    "severity": 2
  }
}
```

</td>
</tr>
</table>

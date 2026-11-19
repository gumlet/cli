---
name: gumlet-account
description: Read current user details, invoice history, and upcoming billing information using the authenticated Gumlet CLI.
---

```sh
gumlet user get
gumlet billing invoice history
gumlet billing invoice upcoming
```

These commands are read-only. Invoice history returns `stripe_invoices` and
the `unpaid_invoices` indicator. Upcoming invoice data may include current
usage and billing dates. Keep account and invoice output out of shared logs.

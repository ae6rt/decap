Go retry package with timeout and retry limits.

Usage

```
...

import "github.com/ae6rt/retry"

retry := retry.New(3*time.Second, 3, retry.DefaultBackoffFunc)
work := func() error {
   // do stuff
   return nil
}

err := retry.Try(work)
if err != nil {
   if retry.IsTimeout(err) {
     fmt.Printf("Timeout\n")
   } else {
     fmt.Printf("Error: %v\n", err)
   }
}
```

Go retry package with timeout and retry limits.

Usage

```
...

import "github.com/ae6rt/retry"

work := func() error {
   // do stuff
   return nil
}

r := retry.New(3, retry.DefaultBackoffFunc)

err := r.Try(work)

if err != nil {
   fmt.Printf("Error: %v\n", err)
}
```

Earlier versions used channel timeouts.  The latest version is
inspired by the much simpler
https://blog.abourget.net/en/2016/01/04/my-favorite-golang-retry-function/.

# Description
This is an implementation of [VLFeat's local binary pattern library](http://www.vlfeat.org/api/lbp.html) written in Go. It implements the local binary patterns algorithm described by (Ojala, Pietikainen, & Harwood, 1996).

# Installation

```go
go get github.com/fastturtle/lbp
```

# Usage

```go
package main

import (
    "github.com/fastturtle/lbp"
    "log"
)

func main() {
    locBinPtrn := lbp.NewLbp()

    feats, err := locBinPtrn.ProcessAll("/path/to/images/folder", "class", 4)
    if err != nil {
        log.Fatal(err)
        return
    }

    moreFeats, err := locBinPtrn.ProcessAll("/more/images", "class", 4)
    if err != nil {
        log.Fatal(err)
        return
    }
    copy(feats, moreFeats)

    evenMoreFeats, err := locBinPtrn.ProcessAll("/even/more/images", "class", 4)
    if err != nil {
        log.Fatal(err)
        return
    }
    copy(feats, evenMoreFeats)

    nn = lbp.NewNearestNeighbor1(feats, 5, lbp.Chisq)

    testFeats, err := locBinPtrn.Process("path/to/test/image", "")
    if err != nil {
        log.Fatal(err)
        return
    }

    class, _ := nn.Classify(testFeats)

    fmt.Println(class)


}
```
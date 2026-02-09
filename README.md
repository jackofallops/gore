# gore - Native Go Regular Expressions (PCRE-style)

`gore` is a powerful, pure Go regular expression library that provides features missing from the standard `regexp` package, such as lookarounds and named capture groups.

It is designed to be a familiar, drop-in replacement for standard regex in many cases, while offering the advanced capabilities of PCRE engines when you need them.

## 🚀 Key Features

*   **Lookarounds**: Supports positive/negative Lookahead `(?=...)`, `(?!...)` and Lookbehind `(?<=...)`, `(?<!...)`.
*   **Named Capture Groups**: Use `(?P<name>...)` to give clarity to your patterns.
*   **Negated Character Classes**: Full support for `[^a-z]` logic.
*   **Streaming Support**: Match patterns directly against `io.Reader` sources (files, network streams) without loading everything into memory.
*   **Pure Go**: No CGo dependencies.

## 📦 Installation

```bash
go get github.com/yourusername/gore
```

## 🛠 Usage

### 1. Drop-in Replacement

You can use `gore` much like the standard `regexp` package.

```go
package main

import (
    "fmt"
    "gore" // or "github.com/..."
)

func main() {
    // Standard matching
    re := gore.MustCompile(`[a-z]+`)
    fmt.Println(re.MatchString("hello world")) // true
}
```

### 2. Advanced Features: Lookarounds

The standard library cannot handle lookarounds. `gore` can.

**Lookahead (Positive & Negative)**

```go
// Match "q" only if followed by "u"
pos := gore.MustCompile(`q(?=u)`)
fmt.Println(pos.MatchString("quit"))  // true
fmt.Println(pos.MatchString("qatar")) // false

// Match "q" only if NOT followed by "u"
neg := gore.MustCompile(`q(?!u)`)
fmt.Println(neg.MatchString("quote")) // false
fmt.Println(neg.MatchString("qatar")) // true
```

**Lookbehind (Positive & Negative)**

```go
// Match "bar" only if preceded by "foo"
behind := gore.MustCompile(`(?<=foo)bar`)
fmt.Println(behind.MatchString("foobar")) // true
fmt.Println(behind.MatchString("bar"))    // false
```

### 3. Named Capture Groups

```go
re := gore.MustCompile(`(?P<first>\w+)\s(?P<last>\w+)`)
// Internal logic tracks names (API for extraction coming soon!)
matched := re.MatchString("John Doe")
if matched {
    // Logic to extract named groups would go here
}
```

### 4. Streaming Support (io.Reader)

Process large inputs efficiently without reading the entire file into memory at once.

```go
file, _ := os.Open("large_log.txt")
re := gore.MustCompile(`ERROR: \d+`)

matched, err := re.MatchReader(file)
if matched {
    fmt.Println("Found error in log file!")
}
```

## ⚠️ Performance Note

Unlike the standard `regexp` package (which uses RE2 and guarantees O(n) linear time), `gore` uses a **backtracking engine** to support these advanced features.

*   **Pros**: Supports lookarounds, backreferences (planned), and complex assertions.
*   **Cons**: Can be slower for certain pathological patterns (exponential time in worst case).

Use `gore` when you need features that `regexp` simply cannot provide. For standard, simple patterns where safety is paramount, the standard library is still a great choice.

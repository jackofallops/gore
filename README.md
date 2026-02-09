# gore - Native Go Regular Expressions (PCRE-style)

DISCLOSURE: This lib was developed as an experiment for PCRE style regexp processing in Go, it was created using Antigravity as a side-project to helpe me skill up in a few areas, and something potentially useful always works best for that. It's provided "as-is" and I make no gaurantees that I'll update or support it in future. That said, I'd love to hear feedback if folks find it useful, or if something definitely doesn't work as expected (I'm no regex expert, and my Perl days are long behind me, thankfully!) 

`gore` is a (hopefully) powerful, pure Go regular expression library that provides features missing from the standard `regexp` package, such as lookarounds and named capture groups.

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

    "github.com/jackofallops/gore"
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

Use `FindStringSubmatch` to extract captures, and `SubexpNames` to map them to names.

```go
re := gore.MustCompile(`(?P<first>\w+)\s(?P<last>\w+)`)
matches := re.FindStringSubmatch("John Doe")
names := re.SubexpNames()

if matches != nil {
    for i, match := range matches {
        if i == 0 { continue } // Skip whole match
        name := names[i]
        if name != "" {
            fmt.Printf("%s: %s\n", name, match)
        } else {
            fmt.Printf("Group %d: %s\n", i, match)
        }
    }
}
// Output:
// first: John
// last: Doe
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

## 🎯 Supported Features

### Character Classes & Escapes
- `[a-z]`, `[^0-9]` - Standard and negated character classes
- `\d`, `\D` - Digits and non-digits
- `\w`, `\W` - Word characters and non-word characters  
- `\s`, `\S` - Whitespace and non-whitespace
- `\n`, `\t`, `\r`, `\f`, `\v` - Literal escapes
- `\b`, `\B` - Word boundaries and non-boundaries

### Quantifiers
- `*`, `+`, `?` - Standard quantifiers (greedy and non-greedy with `?`)
- `{n}` - Exactly n times
- `{n,m}` - Between n and m times
- `{n,}` - n or more times
- All quantifiers support non-greedy variants (e.g., `{2,4}?`)

### Benchmarks (Apple M2)

After extensive optimizations including sync.Pool for allocations, fixed-length lookbehind optimization, and prefix search:

| Benchmark | Time/Op | Memory | Notes |
| :--- | :--- | :--- | :--- |
| `Literal` | ~96 ns | 200 B | Fast prefix search optimization |
| `Lookahead` | ~128 ns | 240 B | Very efficient zero-width assertion |
| `Lookbehind` | ~302 ns | 384 B | Optimized with fixed-length detection |
| `LookbehindLong` | ~66 μs | 64 KB | **227x faster** than naive O(N) with optimization! |
| `Pathological` | ~181 ms | 101 MB | Exponential backtracking, but 85% less memory |
| `NamedCaptures` | ~466 ns | 440 B | Includes capture overhead with pooling |

**Performance Highlights:**
- ✅ Fixed-length lookbehind patterns are **99.6% faster** (15ms → 66μs)
- ✅ 40-85% memory reduction across all patterns vs. baseline
- ✅ Prefix search optimization for literal-heavy patterns

Use `gore` when you need features that `regexp` simply cannot provide. For standard, simple patterns where safety is paramount, the standard library is still a great choice.

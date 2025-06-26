> ⚠️ **This is a fork** of the original [github.com/tmc/langchaingo](https://github.com/tmc/langchaingo) repository.

# 🔗 LangChain Go (fork)

⚡ Building applications with LLMs through composability, with Go! ⚡

## 🚀 Important Announcement 🚀

> **Starting from release [v0.1.13-update.1](https://github.com/sayerxofficial/langchaingo/releases/tag/v0.1.13-update.1)**, this fork is now a full-fledged library that **does not require** any `replace` directives in your go.mod file. You can simply add it to your project with:
>
> ```bash
> go get github.com/sayerxofficial/langchaingo@latest
> ```



## Branch Structure and Versioning

This repository follows a specific branching strategy to maintain both upstream compatibility and custom enhancements:

### Branch Management

- **main**: Fully synchronized with upstream (`tmc/langchaingo`). Never force-pushed.
- **main-pull-requests**: Contains merged PRs from upstream that haven't been officially merged. Rebased on `main` after synchronization (commit hashes will change).
- **main-sayerxofficial**: Default branch containing all current enhancements. Rebased on `main-pull-requests` (commit hashes will change).
- **release/v***: Created from `main-sayerxofficial` for each release. These branches are stable and never force-pushed.

### Versioning

Release tags follow the format `v0.1.13-update.1`, where:
- `v0.1.13` corresponds to the latest upstream release version
- `-update.1` indicates our increment number (starting at 1 and incrementing with each new release)

Each new release cumulatively includes all changes from previous releases on top of the current upstream state, ensuring that you always get the complete set of enhancements when using a specific tag.

### Dependency Management

**Important**: When using this fork in your projects, always reference **release tags** rather than commit hashes. This ensures proper dependency resolution since branches like `main-sayerxofficial` undergo rebasing and their commit hashes change over time.

```
go get github.com/sayerxofficial/langchaingo@v0.1.13-update.1
```

### Branch Visualization

```
  main              A---B---C---D---E---F   (synced with upstream)
                     \
  main-pull-requests   \---G---H---I        (rebased on main, PRs from upstream)
                          \
  main-sayerxofficial           \---J---K---L    (default branch, rebased on main-pull-requests)
                                \
  release/vM.M.P-update.N        M          (tagged stable release)
```

### For Contributors

If you want to contribute to this fork, please create Pull Requests based on the current state of the `main-sayerxofficial` branch. Even though commit hashes in this branch may change due to rebasing, your contributions will be included in the next release when enough changes have accumulated.

When creating a PR, please ensure your changes are well-tested and include appropriate documentation. Once merged, your contributions will be included in the next stable release with fixed commit hashes.




## 🎉 Examples

See [./examples](./examples) for example usage.

```go
package main

import (
  "context"
  "fmt"
  "log"

  "github.com/sayerxofficial/langchaingo/llms"
  "github.com/sayerxofficial/langchaingo/llms/openai"
)

func main() {
  ctx := context.Background()
  llm, err := openai.New()
  if err != nil {
    log.Fatal(err)
  }
  prompt := "What would be a good company name for a company that makes colorful socks?"
  completion, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(completion)
}
```

```shell
$ go run .
Socktastic
```

# Resources


Here are some links to blog posts and articles on using Langchain Go:

- [Using Gemini models in Go with LangChainGo](https://eli.thegreenplace.net/2024/using-gemini-models-in-go-with-langchaingo/) - Jan 2024
- [Using Ollama with LangChainGo](https://eli.thegreenplace.net/2023/using-ollama-with-langchaingo/) - Nov 2023
- [Creating a simple ChatGPT clone with Go](https://sausheong.com/creating-a-simple-chatgpt-clone-with-go-c40b4bec9267?sk=53a2bcf4ce3b0cfae1a4c26897c0deb0) - Aug 2023
- [Creating a ChatGPT Clone that Runs on Your Laptop with Go](https://sausheong.com/creating-a-chatgpt-clone-that-runs-on-your-laptop-with-go-bf9d41f1cf88?sk=05dc67b60fdac6effb1aca84dd2d654e) - Aug 2023




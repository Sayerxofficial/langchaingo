package textsplitter

import (
	"os"
	"testing"

	"github.com/sayerxofficial/langchaingo/schema"

	"github.com/pkoukk/tiktoken-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownHeaderTextSplitter_SplitText(t *testing.T) {
	t.Parallel()

	type testCase struct {
		markdown     string
		expectedDocs []schema.Document
	}

	testCases := []testCase{
		{
			markdown: `
## First header: h2
Some content below the first h2.
## Second header: h2
### Third header: h3

- This is a list item of bullet type.
- This is another list item.

 *Everything* is going according to **plan**.

# Fourth header: h1
Some content below the first h1.
## Fifth header: h2
#### Sixth header: h4

Some content below h1>h2>h4.
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `## First header: h2
Some content below the first h2.`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `## Second header: h2`,
					Metadata:    map[string]any{},
				},
				{
					PageContent: `## Second header: h2
### Third header: h3
- This is a list item of bullet type.`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `## Second header: h2
### Third header: h3
- This is another list item.`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `## Second header: h2
### Third header: h3
*Everything* is going according to **plan**.`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `# Fourth header: h1
Some content below the first h1.`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `# Fourth header: h1
## Fifth header: h2`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `# Fourth header: h1
## Fifth header: h2
#### Sixth header: h4
Some content below h1>h2>h4.`,
					Metadata: map[string]any{},
				},
			},
		},
	}

	splitter := NewMarkdownTextSplitter(WithChunkSize(64), WithChunkOverlap(32), WithHeadingHierarchy(true))
	for _, tc := range testCases {
		docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

// TestMarkdownHeaderTextSplitter_Table markdown always split by line.
//
//nolint:funlen
func TestMarkdownHeaderTextSplitter_Table(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		markdown     string
		options      []Option
		expectedDocs []schema.Document
	}

	testCases := []testCase{
		{
			name: "size(64)-overlap(32)",
			options: []Option{
				WithChunkSize(64),
				WithChunkOverlap(32),
			},
			markdown: `| Syntax      | Description |
| ----------- | ----------- |
| Header      | Title       |
| Paragraph   | Text        |`,
			expectedDocs: []schema.Document{
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Header | Title |`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Paragraph | Text |`,
					Metadata: map[string]any{},
				},
			},
		},
		{
			name: "size(512)-overlap(64)",
			options: []Option{
				WithChunkSize(512),
				WithChunkOverlap(64),
			},
			markdown: `| Syntax      | Description |
| ----------- | ----------- |
| Header      | Title       |
| Paragraph   | Text        |`,
			expectedDocs: []schema.Document{
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Header | Title |`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Paragraph | Text |`,
					Metadata: map[string]any{},
				},
			},
		},
		{
			name: "big-tables-overflow",
			options: []Option{
				WithChunkSize(64),
				WithChunkOverlap(32),
				WithJoinTableRows(true),
			},
			markdown: `| Syntax      | Description |
| ----------- | ----------- |
| Header      | Title       |
| Paragraph   | Text        |`,
			expectedDocs: []schema.Document{
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Header | Title |`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Paragraph | Text |`,
					Metadata: map[string]any{},
				},
			},
		},
		{
			name: "big-tables",
			options: []Option{
				WithChunkSize(128),
				WithChunkOverlap(32),
				WithJoinTableRows(true),
			},
			markdown: `| Syntax      | Description |
| ----------- | ----------- |
| Header      | Title       |
| Paragraph   | Text        |`,
			expectedDocs: []schema.Document{
				{
					PageContent: `| Syntax | Description |
| --- | --- |
| Header | Title |
| Paragraph | Text |`,
					Metadata: map[string]any{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rq := require.New(t)

			splitter := NewMarkdownTextSplitter(tc.options...)

			docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
			rq.NoError(err)
			rq.Equal(tc.expectedDocs, docs)
		})
	}
}

func TestMarkdownHeaderTextSplitter(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("./testdata/example.md")
	if err != nil {
		t.Fatal(err)
	}

	splitter := NewMarkdownTextSplitter(WithChunkSize(512), WithChunkOverlap(64))
	docs, err := CreateDocuments(splitter, []string{string(data)}, nil)
	if err != nil {
		t.Fatal(err)
	}

	var pages string
	for _, doc := range docs {
		pages += doc.PageContent + "\n\n---\n\n"
	}

	err = os.WriteFile("./testdata/example_markdown_header_512.md", []byte(pages), os.ModeExclusive|os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMarkdownHeaderTextSplitter_BulletList(t *testing.T) {
	t.Parallel()
	type testCase struct {
		markdown     string
		expectedDocs []schema.Document
	}
	testCases := []testCase{
		{
			markdown: `
- [Code of Conduct](#code-of-conduct)
- [I Have a Question](#i-have-a-question)
- [I Want To Contribute](#i-want-to-contribute)
    - [Reporting Bugs](#reporting-bugs)
        - [Before Submitting a Bug Report](#before-submitting-a-bug-report)
        - [How Do I Submit a Good Bug Report?](#how-do-i-submit-a-good-bug-report)
    - [Suggesting Enhancements](#suggesting-enhancements)
        - [Before Submitting an Enhancement](#before-submitting-an-enhancement)
        - [How Do I Submit a Good Enhancement Suggestion?](#how-do-i-submit-a-good-enhancement-suggestion)
    - [Your First Code Contribution](#your-first-code-contribution)
        - [Make Changes](#make-changes)
            - [Make changes in the UI](#make-changes-in-the-ui)
            - [Make changes locally](#make-changes-locally)
        - [Commit your update](#commit-your-update)
        - [Pull Request](#pull-request)
        - [Your PR is merged!](#your-pr-is-merged)
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `- [Code of Conduct](#code-of-conduct)
- [I Have a Question](#i-have-a-question)`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `- [I Want To Contribute](#i-want-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
    - [Before Submitting a Bug Report](#before-submitting-a-bug-report)
    - [How Do I Submit a Good Bug Report?](#how-do-i-submit-a-good-bug-report)
  - [Suggesting Enhancements](#suggesting-enhancements)
    - [Before Submitting an Enhancement](#before-submitting-an-enhancement)
    - [How Do I Submit a Good Enhancement Suggestion?](#how-do-i-submit-a-good-enhancement-suggestion)`,
					Metadata: map[string]any{},
				},
				{
					PageContent: `  - [Your First Code Contribution](#your-first-code-contribution)
    - [Make Changes](#make-changes)
      - [Make changes in the UI](#make-changes-in-the-ui)
      - [Make changes locally](#make-changes-locally)
    - [Commit your update](#commit-your-update)
    - [Pull Request](#pull-request)
    - [Your PR is merged!](#your-pr-is-merged)`,
					Metadata: map[string]any{},
				},
			},
		},
	}

	for _, tc := range testCases {
		splitter := NewMarkdownTextSplitter(WithChunkSize(512), WithChunkOverlap(64))
		docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

func TestMarkdownHeaderTextSplitter_HeaderAfterHeader(t *testing.T) {
	t.Parallel()

	type testCase struct {
		markdown     string
		expectedDocs []schema.Document
	}

	testCases := []testCase{
		{
			markdown: `
### Your First Code Contribution

#### Make Changes

##### Make changes in the UI

Click **Make a contribution** at the bottom of any docs page to make small changes such as a typo, sentence fix, or a
broken link. This takes you to the .md file where you can make your changes and [create a pull request](#pull-request)
for a review.

##### Make changes locally

1. Fork the repository.

2. Install or make sure **Golang** is updated.

3. Create a working branch and start with your changes!
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `### Your First Code Contribution`, Metadata: map[string]any{},
				},
				{
					PageContent: `#### Make Changes`, Metadata: map[string]any{},
				},
				{
					PageContent: `##### Make changes in the UI
Click **Make a contribution** at the bottom of any docs page to make small changes such as a typo, sentence fix, or a
broken link. This takes you to the .md file where you can make your changes and [create a pull request](#pull-request)
for a review.`, Metadata: map[string]any{},
				},
				{
					PageContent: `##### Make changes locally
1. Fork the repository.
2. Install or make sure **Golang** is updated.
3. Create a working branch and start with your changes!`, Metadata: map[string]any{},
				},
			},
		},
	}

	for _, tc := range testCases {
		splitter := NewMarkdownTextSplitter(WithChunkSize(512), WithChunkOverlap(64))
		docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

//nolint:funlen
func TestMarkdownHeaderTextSplitter_SplitCode(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		options      []Option
		markdown     string
		expectedDocs []schema.Document
	}{
		{
			name:     "fence-false",
			markdown: "example code:\n```go\nfunc main() {}\n```",
			expectedDocs: []schema.Document{
				{
					PageContent: "example code:",
					Metadata:    map[string]any{},
				},
			},
		},
		{
			name: "fence-true",
			options: []Option{
				WithCodeBlocks(true),
			},
			markdown: "example code:\n```go\nfunc main() {}\n```",
			expectedDocs: []schema.Document{
				{
					PageContent: "example code:\n\n```go\nfunc main() {}\n```\n",
					Metadata:    map[string]any{},
				},
			},
		},
		{
			name: "codeblock-false",
			markdown: `example code:

    func main() {
	}
    `,
			expectedDocs: []schema.Document{
				{
					PageContent: `example code:`,
					Metadata:    map[string]any{},
				},
			},
		},
		{
			name: "codeblock-true",
			options: []Option{
				WithCodeBlocks(true),
			},
			markdown: `example code:

    func main() {
	}
    `,
			expectedDocs: []schema.Document{
				{
					PageContent: `example code:

    func main() {
    }
    `,
					Metadata: map[string]any{},
				},
			},
		},
		{
			name: "hr",
			markdown: `example code:

---
more text
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `example code:

---
more text`,
					Metadata: map[string]any{},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rq := require.New(t)

			splitter := NewMarkdownTextSplitter(append(tc.options,
				WithChunkSize(512),
				WithChunkOverlap(64),
			)...)

			docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
			rq.NoError(err)
			rq.Equal(tc.expectedDocs, docs)
		})
	}
}

//nolint:funlen
func TestMarkdownHeaderTextSplitter_SplitInline(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name         string
		options      []Option
		markdown     string
		expectedDocs []schema.Document
	}{
		{
			name:     "break",
			markdown: "text with\\\nhard break\nsoft break",
			expectedDocs: []schema.Document{
				{
					PageContent: "text with\\\nhard break\nsoft break",
					Metadata:    map[string]any{},
				},
			},
		},
		{
			name:     "emphasis",
			markdown: "text with *emphasis*, **strong emphasis** and ~~strikethrough~~",
			expectedDocs: []schema.Document{
				{
					PageContent: "text with *emphasis*, **strong emphasis** and ~~strikethrough~~",
					Metadata:    map[string]any{},
				},
			},
		},
		{
			name: "image",
			markdown: `images:
![one](/path/to/one.png)
![two](/path/to/two.png "two")
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `images:
![one](/path/to/one.png)
![two](/path/to/two.png "two")`,
					Metadata: map[string]any{},
				},
			},
		},
		{
			name: "link-false",
			markdown: `links:
[foo][bar]

[bar]: /url "title"

[regular](/url)
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `links:
[foo][bar]
[regular](/url)`,
					Metadata: map[string]any{},
				},
			},
		},
		{
			name: "link-true",
			options: []Option{
				WithReferenceLinks(true),
			},
			markdown: `links:
[foo][bar]

[bar]: /url "title"

[regular](/url)
`,
			expectedDocs: []schema.Document{
				{
					PageContent: `links:
[foo](/url "title")
[regular](/url)`,
					Metadata: map[string]any{},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rq := require.New(t)

			splitter := NewMarkdownTextSplitter(append(tc.options,
				WithChunkSize(512),
				WithChunkOverlap(64),
			)...)

			docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
			rq.NoError(err)
			rq.Equal(tc.expectedDocs, docs)
		})
	}
}

func TestMarkdownHeaderTextSplitter_LenFunc(t *testing.T) {
	t.Parallel()

	tokenEncoder, _ := tiktoken.GetEncoding("cl100k_base")

	sampleText := "The quick brown fox jumped over the lazy dog."
	tokensPerChunk := len(tokenEncoder.Encode(sampleText, nil, nil))

	type testCase struct {
		markdown     string
		expectedDocs []schema.Document
	}

	testCases := []testCase{
		{
			markdown: `# Title` + "\n" + sampleText + "\n" + sampleText,
			expectedDocs: []schema.Document{
				{
					PageContent: "# Title" + "\n" + sampleText,
					Metadata:    map[string]any{},
				},
				{
					PageContent: "# Title" + "\n" + sampleText,
					Metadata:    map[string]any{},
				},
			},
		},
	}

	splitter := NewMarkdownTextSplitter(
		WithChunkSize(tokensPerChunk+1),
		WithChunkOverlap(0),
		WithLenFunc(func(s string) int {
			return len(tokenEncoder.Encode(s, nil, nil))
		}),
	)

	for _, tc := range testCases {
		docs, err := CreateDocuments(splitter, []string{tc.markdown}, nil)
		require.NoError(t, err)
		assert.Equal(t, tc.expectedDocs, docs)
	}
}

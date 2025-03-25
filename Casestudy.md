---
title: Creating simple tools with the help of Claude
date: 2023-04-15
tags: Go, AI, VibeCoding, DevTooling
---

# Creating simple tools with the help of Claude

I used Claude to generate a simple single page program to help take some of the tedium out of writing go code.
Go has verbose error handling with some very common cases. I wanted to create a simpler writing experience for those cases.

## Challenge

So often a go programmer will write something in the following form

```go
err := someFunction()
if err != nil {
  return err
}
```

That or they'll need to include some zero values to match the function return types

```go
err := otherFunction()
if err != nil {
  return 0, "", err
}
```

I wanted to simplify the writing experience of go with a simpler, syntactic sugar for these cases.

```go
_ := someFunction //?
x, y, _ := otherFunction //? 0, "", err
```

## Solution

I prompted Claude with a detailed specification of the behavior I wanted, with examples of the desugaring process.
I had to get Claude to simplify their solution as they ran into a string of errors they were clearly having trouble resolving, but within an hour or so I ended up with the solution seen here.

## Results

This is a preprocessor that improves the writing experience for some common Go error handling situations while remaining valid Go code. Future work could include IDE integration to desugar as an engineer works.

## Technologies Used

Golang
Claude

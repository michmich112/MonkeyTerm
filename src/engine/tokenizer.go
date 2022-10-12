package engine

type Tokenizer interface {
  // Seeds the tokenizer
  Seed(seed int64) 

  // Gets the next token
  Next() (token string, eof bool)
}

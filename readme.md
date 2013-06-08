Sort-of-almost-finished Go identifier renaming. If it were finished, it would assume that the declaration of the variable to be found is the first statement on its line, and that the source is syntactically correct.

What's left to do:

 - Find the block the scope of which contains the identifier. (getnode.go)
 - Do the I/O part. (main.go)
 - Support renaming of struct fields and interface methods. (Selectors are tricky.)

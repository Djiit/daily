# Edge Cases for Task Parsing

## Malformed Tasks
- [ Incomplete task syntax
- [x Missing closing bracket
- [] Empty task checkbox
- [y] Invalid checkbox character

## Tasks in Code Blocks
```markdown
- [ ] This task is in a code block and should not be parsed
- [x] This one too
```

## Tasks in Quotes
> - [ ] This task is in a blockquote
> - [x] Completed task in blockquote

## Empty Lines and Spacing
- [ ] Task with normal spacing

- [ ] Task with extra line above

- [ ]    Task with extra spaces
- [x]Task with no space after checkbox

## Very Long Tasks
- [ ] This is a very long task description that goes on and on and might test how the parser handles really long task descriptions that could potentially cause issues with formatting or storage in the system when displayed or processed by various components of the application

## Tasks with Special Characters
- [ ] Task with émojis 🚀 and spëcial chars áéíóú
- [ ] Task with "quotes" and 'apostrophes'
- [ ] Task with [brackets] and (parentheses)
- [ ] Task with <angle brackets> and {curly braces}

## Mixed List Types
1. [ ] Numbered task item
2. [x] Completed numbered task
   - [ ] Nested task under numbered item
   - [x] Nested completed task

* [ ] Bullet task with asterisk
* [x] Completed bullet task with asterisk
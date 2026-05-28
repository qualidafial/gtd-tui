## 1. Model changes

- [x] 1.1 Add `projectName string` field to `taskedit.Model` and update `New()` signature to accept it
- [x] 1.2 Render "Project" meta line in `View()` when projectName is non-empty

## 2. Caller updates

- [x] 2.1 Update call sites that construct `taskedit.New(...)` to resolve project name from ProjectID via ProjectService before opening the editor

## 3. Tests

- [x] 3.1 Update existing taskedit model tests for the new signature
- [x] 3.2 Add test case verifying Project line appears when projectName is set
- [x] 3.3 Add test case verifying Project line is absent when projectName is empty
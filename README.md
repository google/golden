Package golden provides a function for comparing arbitrary text data to
a golden file, and allowing a person running a test to automatically set the
golden data to the "actual" data by setting a flag when running the unit
test.

This library is most useful when the data being compared is large, and
correctness can be easily determined by having a human look at a diff.  Some
examples are the output of an HTML template, or the output of a large
computation formatted as a text protobuf.

Expected usage:

```go
func AUnitTest(t *testing.T) {
  got := proto.MarshalTextString(code_under_test.ComputeTediousData(...))
  if diff := golden.Compare(got, ".../testdata/data.txt.golden"); diff != "" {
    t.Error(diff)
  }
}
```

When the user runs this test and the actual data differs from the golden
file, they will see the following error message:

```diff
Actual data differs from golden data; run "go test -update_golden" to update
--- .../testdata/data.txt.golden
+++ .../testdata/data.txt.actual
  blah: ""
- ultimate_answer: 41
+ ultimate_answer: 42
  foo: "bar"
  baz: "blah"
```

The user will inspect the diff. Let's say that this diff is due to a simple
code change, and the new value is definitely correct. The user can then
decide to overwrite the golden data with the actual data by re-running the
test, passing the `-update_golden` flag:

```
$ go test -update_golden
```

This time the test will succeed, and the golden data will be overwritten
with the actual data. Code reviewers will notice in diffs that the golden
data has been modified, and can easily compare the output of the code before
and after the change.

This is not an official Google product.

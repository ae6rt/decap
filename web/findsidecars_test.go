package main

/*
func TestFindSidecars(t *testing.T) {
	dir, err := ziptools.Unzip("buildscripts-repo.zip")
	if err != nil {
		t.Fatal(err)
	}

	scripts, err := findBuildScripts(dir)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}

	// Sidecars are located on-par with a build.sh
	for _, v := range scripts {
		parent := parentPath(v)
		sidecars, err := findSidecars(parent)
		if err != nil {
			os.RemoveAll(dir)
			t.Fatal(err)
		}
		if strings.HasSuffix(parent, "ae6rt/dynamodb-lab") && len(sidecars) != 1 {
			os.RemoveAll(dir)
			t.Fatalf("Want 1 but got %d\n", len(sidecars))
		}
	}
	os.RemoveAll(dir)
}
*/

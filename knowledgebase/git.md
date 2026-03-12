# Revert Part of a Commit

```
$ git revert -n $bad_commit    # Revert the commit, but don't commit the changes
$ git reset HEAD .             # Unstage the changes
$ git add --patch .            # Add whatever changes you want
$ git commit                   # Commit those changes
```

or

```
git revert --no-commit <commit hash>
git reset -p        # every time choose 'y' if you want keep the change, otherwise choose 'n'
git commit -m "Revert ..."
git checkout -- .   # Don't forget to use it.
```


Git Commands to Partially Revert the Last Commit
Here are the commands to revert specific files from your most recent commit:

1. Revert the files:

```

git checkout HEAD^ -- file/path/one.txt file/path/two.js

```

  • This command pulls the content of the specified files from the commit before the current one (`HEAD^`) into your working directory and staging area, effectively undoing the changes from the latest commit for those files.

  • Remember to replace `file/path/one.txt` and `file/path/two.js` with the actual paths to the files you want to revert.

2. Commit the revert:

```

git commit -m "Revert specific files from last commit"

```

  • This creates a new commit that records the undoing of the changes for those specific files.

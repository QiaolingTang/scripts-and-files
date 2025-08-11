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

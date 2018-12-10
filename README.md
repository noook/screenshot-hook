# Screenshot Hook

## Context

For now, we are currently unable to take a screenshot and upload it directly on a server. Not with native apps though.

The goal is to achieve the combination of making a screenshot, saving it, upload the screenshot onto a server, remove the temporary file, and copying the link to the screenshot into the clipboard to ease sharing.

## Problematic

How can we achieve all these actions in a single script ? And why not choosing an alternative solution like Monosnap ?

## Concept

+ Enable to run the script from a keyboard shortcut
+ Generating a short and unique ID for the screenshot
+ Transfering through SSH the file onto the server
+ Copying the link to the image in the clipboard
+ Add configuration in a `.screenshot-hookrc` file at the root of a user's directory
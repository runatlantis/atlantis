# Glossary

The Atlantis community uses many words and phrases to work more efficiently.
You will find the most common ones and their meaning on this page.

## Pull / Merge Request Event

The different VCSs have different names for merging changes. Atlantis uses the
name Pull Request as the abstraction. The VCS provider implements this
abstraction and forwards the call to the respective function.

## VCS

VCS stands for Version Control System.

Atlantis supports different VCSs. Each VCS requires a custom implementation that
abstracts the Atlantis functionalities to the specific VCS implementations.

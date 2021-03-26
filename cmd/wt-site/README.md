# SiteConfig
The **site.thtml** file determines the structure of a static site, and is needed by the **wt-site** utility

It contains several sections:
* *pages*
  * key is dst html file: value is src thtml file, or list of src thtml file with parameters
* *scripts*
  * key is src tjs script: value is dst html file, or list of dst html files
  * multiple scripts can be applied to each view (which are all smartly loaded)
* *styles*
  * key is src thtml file: value is dst html file, or list of dst html files
* *files*
  * files that are simply copied
  * key is dst file: value is src file
* *search*
  * key is dst file: value is dict (SearchConfig)
    * *pages*: list of dst html files, or empty to select all
    * *ignore*: list of words to ignore
    * *title*: css query to find title, required
    * *content*: css query to find content, not required (defaults to all text except title)

The SiteConfig is only needed by wt-site
SearchConfig is also needed by wt-search
File-extensions are completely optional (but it is highly recommended to adhere to standards)

# Automatically generated files
* bundle.js (whichever name is unique after processing of *files*)
* style0.css, style1.css, ... (whichever name is unique after processing of *files*)
* math.woff2 (whichever name is unique after processing of *files*)

# Cache
Should be technology agnostic

A map of dst to dependencies (i.e. sources)
Eg. for files this is simply one-on-one.
Also an env string determines if everything requires a rebuild or not (if different)

* A change is defined as being more recent than any of its dependencies
* Perhaps that lastModified time should be stored in the gob?
* If a source file changes, then all destinations that depend on it are deleted from the cache
* Destinations that are no longer in the cache should be deleted
Because it is based 
* What if style0.css path changes?, so each each dst should have some build parameters
* Each dst search index

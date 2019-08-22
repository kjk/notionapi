@cd do
@go run . %*

@rem notable pages:
@rem 0367c2db381a4f8b9ce360f388a6b2e3 : my test pages
@rem d6eb49cfc68f402881af3aef391443e6 : Pokedex
@rem 3b617da409454a52bc3a920ba8832bf7 : Blendle's Employee Handbook

@rem notable flags
@rem -test-to-html ${pageID} : compares html rendering of page with Notion
@rem -test-to-md ${pageID} : c
@rem -to-html ${pageID} : downloads page, converts to html and
@rem -re-export : for -test-to-html and -test-to-md, re-export latest version from Notion
@rem -no-cache : disables download cache

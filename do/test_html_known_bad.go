package main

var knownBadHTML = [][]string{
	{
		// id of the start page
		"3b617da409454a52bc3a920ba8832bf7",

		// Notion is missing one link to page
		"13aa42a5a95d4357aa830c3e7ff35ae1",
		// TODO(1): Notion renders <div> inside <p> which is illegal and makes pretty-printing
		// not work, so can't compare results. Probably because they render children
		//  <p> inside </p> instead of after </p>
		"4f5ee5cf485048468db8dfbf5924409c",
		// Notion is missing one link to page
		"7a5df17b32e84686ae33bf01fa367da9",
		// Notion is malformed
		"7afdcc4fbede49bc9582469ad6e86fd3",
		// Notion is malformed
		"949f33cdba814fc4a288d81c6e7c810d",
		// Notion is missing one link to page
		"b1b31f6d3405466c988676f996ce03ad",
		// Notion is missong some link to page
		"ef850413bb53491eaebccaa09eeb8630",
		// Notion is malformed
		"f2d97c9cba804583838acf5d571313f5",
		// Notion is malformed
		"3c892714f4dc4d2194619fdccba48fc6",
		// Different ids
		"8f12cc5182a6437aac4dc518cb28b681",
	},
	{
		"0367c2db381a4f8b9ce360f388a6b2e3",

		// TODO: Notion doen't export link to page
		"86b5223576104fa69dc03675e44571b7",
		// TODO: a date with time zone not formatted correctly
		"97100f9c17324fd7ba3d3c5f1832104d",
		// TODO: bad indent in toc
		"c969c9455d7c4dd79c7f860f3ace6429",
		// TODO: Notion exports a column "Title" marked as "not visible"
		"92dd7aedf1bb4121aaa8986735df3d13",
		// TODO: don't have name of the page
		"f97ffca91f8949b48004999df34ab1f7",
	},
	{
		"d6eb49cfc68f402881af3aef391443e6",

		// TODO: I'm not formatting table correctly
		"00f68316d03c4830b00c453e542a1df7",
		// TODO: I'm not formatting table correctly
		"02bfca37eae5484ba942a00c99076b7a",
		// TODO: I'm not formatting table correctly
		"09e9c8f5c9df445f94d1cf3f39a1039f",
		// TODO: totally different export
		"0e684b2e45ea434293274c802b5ad702",
		// TODO: I'm not exporting a table the right way
		"141c2ef1718b471896c915ae622dae83",
		// TODO: Bad export
		"14d22d99fb074352a59d78751646cf3d",
	},
}

func findKnownBadHTML(pageID string) []string {
	for _, a := range knownBadHTML {
		if a[0] == pageID {
			return a[1:]
		}
	}
	return nil
}

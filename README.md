javadocset
==========

This is a port of Kapeli's javadocset tool in Golang. I ported this tool as an exercise for me to use Golang and also to contribute to Zealdoc's users as most Dash users who use a Mac are able to run Kapeli's javadocset tool (which can be found [here](https://github.com/Kapeli/javadocset)).

# Build Instructions

## Prerequisites
You need to have Go 1.8.3 and above to be able to compile this tool

## Instructions
```
// Clone this repo
git clone https://github.com/william8th/javadocset

// Change into the repo's directory
cd javadocset

// Build the tool
go build

// Run the javadocset tool
./javadocset

```

## Usage
```
./javadocset <docset name> <Javadoc folder>

// where:
// 	<docset name> = the name of the docset that will appear on Dash/Zeal
// 	<Javadoc folder> = the folder containing the generated HTML Java doc 
```

# Credits
Credits go to Kapeli: [https://github.com/Kapeli/javadocset](https://github.com/Kapeli/javadocset)

Zeal: [https://github.com/zealdocs/zeal](https://github.com/zealdocs/zeal)

Dash: [https://kapeli.com/dash](https://kapeli.com/dash)

## Note
There is no guarantee that this will work 100%. I've tried my best to verify against a Java doc that I own but there may be bugs that I've not had the chance to discover. Please open an issue if you're having any trouble!

Pull requests are welcome :)

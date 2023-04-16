# Porosity-Calculator
A golang tool created to calculate porosity for a Materials Design Capstone Project

## Installation

First, ensure you have Go installed on your machine. You can find instructions for your operating system [here](https://go.dev/doc/install).

Next, clone this repository to your machine. You can do this by running the following command in your terminal:

```bash
git clone https://github.com/VINXIS/Porosity-Calculator.git
```

Once you have cloned the repository, navigate to the directory, and create a folder called `original`. This is where your optical microscopy images would be placed.
    
```bash
    cd Porosity-Calculator
    mkdir original
```

Once you have the `original` folder, you can place your images in there. 
The images should be in the `x-yD.png` format, where x is the sample, y is the image iteration, and D is the direction you are going to take the images. 
You can then run the following command to calculate the porosity of the images:

```bash
go run core.go
```

This will create a new folder called `processed` in the directory. This folder will contain the processed images for every b value threshold, as well as a file called `porosity.csv` that contains the porosity of each image at every b value. This file will contain the porosity of each image.
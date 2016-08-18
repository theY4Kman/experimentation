# Genetic Arithmetic Expression Solver

Execute `python genetic_expr_ui.py` to get a graphical interface. 

Description: a number is randomly chosen, and the aim of the program is to drum up a simple arithmetic equation using only addition, subtraction, multiplication, and division of a number of components less than 10. For instance, for the number 184, a solution might be `6*7*4+1-5+8+7+5`.

To find its solutions, the program begins with a population of 30 "chromosomes", each consisting of 30 "genes" representing digits or operators. At each step, two chromosomes are chosen at random-ish. We roll a die to see if we should crossover them, which is to swap their bits around a random bit

The program defines a simple language with characters defined as 4-bit nibbles:

| Index  | Binary | Character |
| ------:| ------ | --------- |
|    `0` | `0000` | 0         |
|    `1` | `0001` | 1         |
|    `2` | `0010` | 2         |
|    `3` | `0011` | 3         |
|    `4` | `0100` | 4         |
|    `5` | `0101` | 5         |
|    `6` | `0110` | 6         |
|    `7` | `0111` | 7         |
|    `8` | `1000` | 8         |
|    `9` | `1001` | 9         |
|   `10` | `1010` | +         |
|   `11` | `1011` | -         |
|   `12` | `1100` | *         |
|   `13` | `1101` | /         |
|   `14` | `1110` | Invalid   |
|   `15` | `1111` | Invalid   |

These are called genes. A chromosome is a list of genes. In the demonstration, we use chromosomes with 30 genes.


TODO

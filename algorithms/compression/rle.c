/**
 * =============================================================================
 * Run Length Encoding
 * Copyright (C) 2009 Zach "theY4Kman" Kanzler
 * =============================================================================
 *
 * This program is free software; you can redistribute it and/or modify it under
 * the terms of the GNU General Public License, version 3.0, as published by the
 * Free Software Foundation.
 * 
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 * FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
 * details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program.  If not, see <http://www.gnu.org/licenses/>.
 */

#include <stdio.h>
#include <malloc.h>

#define BLOCK_SIZE 255

/**
 * @brief   Compresses a block of data using run length encoding
 *
 * @param in    The input stream to compress
 * @param size  The size of the input stream
 * @param out   A place to save the output stream
 * @return  The size of the compressed stream on success, 0 on failure. For an
 *          input stream size of 0, this will still return 0, as it's an error
 *          to compress nothing.
 */
unsigned int
rle_compress(FILE *in, FILE *out)
{
    unsigned int compressed_size;
    unsigned int buffer_size;
    unsigned char *buffer;
    unsigned int cur_pos;
    unsigned int out_pos;
    
    compressed_size = 0;
    cur_pos = 0;
    
    /* A buffer to store data from the input file */
    buffer = malloc(BLOCK_SIZE);
    if (buffer == NULL)
    {
        fprintf(stderr, "Unable to allocate %d bytes of memory\n", BLOCK_SIZE);
        return 0;
    }
    buffer_size = BLOCK_SIZE;
    
    while (!feof(in))
    {
        unsigned char limit = fread(buffer, sizeof(unsigned char), BLOCK_SIZE,
            in);
        
        unsigned char initial_byte = buffer[cur_pos];
        unsigned char repeated = 1;
        for (unsigned char rep_pos=cur_pos; repeated<limit; rep_pos++)
        {
            if (rep_pos + cur_pos >= BLOCK_SIZE)
            {
                
            }
            
            if (buffer[cur_pos + rep_pos] != initial_byte)
                break;
            repeated++;
        }
        
        fputc(out, repeated);
        fputc(out, initial_byte);
        
        cur_pos++;
    }
    
    *out = buffer;
    
    return compressed_size;
}

int
main (int argc, char *argv[])
{
    switch (argc)
    {
        case 1:
            fprintf(stderr, "No input or output file specified\n");
            return 2;
        
        case 2:
            fprintf(stderr, "No output file specified\n");
            return 2;
        
        default:
            break;
    }
    
    FILE *input_file = fopen(argv[1], "rb");
    if (input_file == NULL)
    {
        fprintf(stderr, "Error opening the input file\n");
        return 2;
    }
    
    FILE *output_file = fopen(argv[2], "wb");
    if (output_file == NULL)
    {
        fprintf(stderr, "Error opening the output file\n");
        return 2;
    }
    
    fseek(input_file, 0, SEEK_END);
    int input_length = ftell(input_file);
    fseek(input_file, 0, SEEK_SET);
    
    unsigned char *out;
    unsigned int compressed_length = rle_compress(input_file, input_length,
        output_file);
    
    printf("Compressed \"%s\" from %d bytes to %d bytes\n", argv[1],
        input_length, compressed_length);
    
    return 0;
}

#ifndef RESIZE_H
#define RESIZE_H

#include <sys/types.h>

/**
 * takes in an image blob, length and the dimensions to resize.
 * Returns a resized blob.
 * NOTE: this does not free any memory
 */
void * resize_image(void * blob, size_t *length, int width, int height,
    int quality, int filter, double blur);

// calls InitializeMagick
void CreateMagick();

// calls DestroyMagick
void DestroyMagick();

#endif /* end of include guard: RESIZE_H */

#ifndef RESIZE_H
#define RESIZE_H

#include <magick/api.h>
#include <sys/types.h>

/**
 * takes in an image blob, length and the dimensions to resize.
 * Returns a resized blob.
 * NOTE: this does not free any memory
 */
void * resize_image(void * blob, size_t *length, int width, int height,
    int quality, int filter, double blur);

typedef enum {
  CAP_IMAGE_ERROR_OK = 0,
  CAP_IMAGE_ERROR_UNEXPECTED_NULL,
  CAP_IMAGE_ERROR_BAD_IMAGE,
  CAP_IMAGE_ERROR_RESIZE_FAIL,
  CAP_IMAGE_ERROR_TOO_LARGE
} cap_image_error;

typedef enum {
  CAP = 0,
  WIDTH,
  HEIGHT
} dimension_enum;

/**
 * If this returns NULL, error will be set.
 * maxWidth or maxHeight of zero means no limit.
 */
void * cap_image(void * blob,
    size_t * length,
    cap_image_error * error,
    int cap,
    dimension_enum dimension,
    int quality,
    int filter,
    double blur,
    int maxWidth,
    int maxHeight);


#endif /* end of include guard: RESIZE_H */

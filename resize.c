#include <stdbool.h>
#include <stdio.h>

#include "resize.h"

int DefaultImageQuality = 75;

void rotate_image(Image ** image, ExceptionInfo * exception) {
  Image * temp_image = NULL;
  switch ((*image)->orientation) {
    case BottomRightOrientation:
      temp_image = RotateImage(*image, 180, exception);
      break;
    case RightTopOrientation:
      temp_image = RotateImage(*image, 90, exception);
      break;
    case LeftBottomOrientation:
      temp_image = RotateImage(*image, 270, exception);
      break;
    default:
      break;
  }
  if (temp_image != (Image *) NULL) {
    DestroyImage(*image);
    *image = temp_image;
  }

  // if we rotated an image with Orientation set, clients will apply the transform again so need to clear EXIF
  StripImage(*image);
}

void * resize_image(void * blob, size_t *length, int width, int height,
    int quality, int filter, double blur)
{
  Image * image = (Image *) NULL;

  void * out_blob = NULL;

  ImageInfo * image_info = CloneImageInfo((ImageInfo *) NULL);
  image_info->quality = quality > 0 ? quality : DefaultImageQuality;

  // get the gm exception buffer set up
  ExceptionInfo exception;
  GetExceptionInfo(&exception);

  if (!blob || !length) {
    goto end;
  }

  image = BlobToImage(image_info, blob, *length, &exception);

  // 0 out length in case we fail at some point.
  *length = 0;

  // If there was an exception or no image was found, give up
  if (exception.severity != UndefinedException || image == (Image *) NULL) {
    CatchException(&exception);
    goto end;
  }

  rotate_image(&image, &exception);

  unsigned int srcWidth = image->columns;
  unsigned int srcHeight = image->rows;

  unsigned int resizeWidth, resizeHeight;

  if (width * srcHeight < height * srcWidth) {
    // Note that resizeWidth > width
    resizeWidth = (height * srcWidth) / srcHeight;
    resizeHeight = height;
  } else {
    // Note that resizeHeight > width
    resizeWidth = width;
    resizeHeight = (width * srcHeight) / srcWidth;
  }

  // resize the image
  // for what its worth, this will totally mess up gifs. gifs need to be
  // coalasced first so we can correctly interpolate each frame.
  Image * temp_image = ResizeImage(image, resizeWidth, resizeHeight, (FilterTypes) filter, blur,  &exception);

  // if the resize failed, return NULL
  if (exception.severity != UndefinedException || temp_image == (Image *) NULL) {
    CatchException(&exception);
    goto end;
  }
  else {
    DestroyImage(image);
    image = temp_image;
  }

  unsigned int cropX = resizeWidth - width;
  unsigned int cropY = resizeHeight - height;

  if (cropX != 0 || cropY != 0) {
    // We crop from the edges.
    RectangleInfo geometry;
    Image * temp_image = (Image *) NULL;
    geometry.width = width;
    geometry.height = height;
    geometry.x = cropX / 2;
    geometry.y = cropY / 2;
    temp_image = CropImage(image, &geometry, &exception);

    if (exception.severity != UndefinedException || temp_image == (Image *) NULL) {
      CatchException(&exception);
      goto end;
    }
    else {
      DestroyImage(image);
      image = temp_image;
    }
  }

  // push the image to a blob
  out_blob = ImageToBlob(image_info, image, length, &exception);

end:

  if (image)
    DestroyImage(image);

  DestroyImageInfo(image_info);
  DestroyExceptionInfo(&exception);

  return out_blob;
}

void * cap_image(void * blob, size_t * length, cap_image_error * error,
    int cap, cap_dimension dimension, int quality, int filter, double blur, int maxWidth, int maxHeight)
{
  Image * image = (Image *) NULL;

  void * out_blob = NULL;
  *error = CAP_IMAGE_ERROR_OK;

  ImageInfo * image_info = CloneImageInfo((ImageInfo *) NULL);
  image_info->quality = quality > 0 ? quality : DefaultImageQuality;

  // get the gm exception buffer set up
  ExceptionInfo exception;
  GetExceptionInfo(&exception);

  if (!blob || !length) {
    *error = CAP_IMAGE_ERROR_UNEXPECTED_NULL;
    goto end;
  }

  image = BlobToImage(image_info, blob, *length, &exception);

  // 0 out length in case we fail at some point.
  *length = 0;

  // If no image was found, give up
  if (exception.severity != UndefinedException || image == (Image *) NULL) {
    CatchException(&exception);
    *error = CAP_IMAGE_ERROR_BAD_IMAGE;
    goto end;
  }

  rotate_image(&image, &exception);

  unsigned int srcWidth = image->columns;
  unsigned int srcHeight = image->rows;

  unsigned int resizeWidth, resizeHeight;

  // Note: resize to max dimension or whatever dimension is requested
  if (dimension == HEIGHT || (dimension != WIDTH && srcWidth < srcHeight)) {
    // force cap on height
    resizeWidth = (cap * srcWidth) / srcHeight;
    resizeHeight = cap;
  } else {
    // force cap on width
    resizeWidth = cap;
    resizeHeight = (cap * srcHeight) / srcWidth;
  }

  if (((maxWidth != 0) && (resizeWidth > (unsigned int)maxWidth)) ||
      ((maxHeight != 0 && resizeHeight > (unsigned int)maxHeight))) {
    *error = CAP_IMAGE_ERROR_TOO_LARGE;
    goto end;
  }

  // Note that here resizeWidth <= cap and resizeHeight <= cap and at least one of them is equal.
  // This means this is the largest size we can have that still adheres to the cap.
  Image * temp_image = ResizeImage(image, resizeWidth, resizeHeight, (FilterTypes) filter, blur,  &exception);

  // if the resize failed, return NULL
  if (exception.severity != UndefinedException || temp_image == (Image *) NULL) {
    CatchException(&exception);
    *error = CAP_IMAGE_ERROR_RESIZE_FAIL;
    goto end;
  }
  else {
    DestroyImage(image);
    image = temp_image;
  }

  // push the image to a blob
  out_blob = ImageToBlob(image_info, image, length, &exception);

end:

  if (image)
    DestroyImage(image);

  DestroyImageInfo(image_info);
  DestroyExceptionInfo(&exception);

  return out_blob;

}

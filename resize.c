#include <magick/api.h>
#include <stdbool.h>

#include "resize.h"

#define WIDTH   1
#define HEIGHT  2
int DefaultImageQuality = 75;

void CreateMagick() {
  InitializeMagick(NULL);
}

void DestroyMagick() {
  DestroyMagick();
}

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

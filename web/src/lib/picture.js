const SIZE = 200;
const SCAN = 480;

/** A product's picture URL, or null. A replaced picture has a new version, so a new URL. */
export const pictureUrl = (product) =>
  product.imageVersion ? `/api/images/${product.id}?v=${product.imageVersion}` : null;

// The box around everything that is neither transparent nor white.
function contentBox(context, width, height) {
  const { data } = context.getImageData(0, 0, width, height);
  let left = width;
  let right = -1;
  let top = height;
  let bottom = -1;
  for (let y = 0; y < height; y++) {
    for (let x = 0; x < width; x++) {
      const i = (y * width + x) * 4;
      if (data[i + 3] < 16 || Math.min(data[i], data[i + 1], data[i + 2]) > 244) continue;
      if (x < left) left = x;
      if (x > right) right = x;
      if (y < top) top = y;
      if (y > bottom) bottom = y;
    }
  }
  if (right < 0) return { left: 0, top: 0, width, height };
  return { left, top, width: right - left + 1, height: bottom - top + 1 };
}

/**
 * Turns any picture into a small WebP before it leaves the browser, so a 5 MB phone photo costs
 * the till about 10 KB. White edges are cut off and the shape stays, so bottles and long
 * chocolate bars both fill their key.
 */
export async function shrinkPicture(file) {
  const bitmap = await createImageBitmap(file);
  const fit = Math.min(SCAN / bitmap.width, SCAN / bitmap.height, 1);
  const scan = document.createElement('canvas');
  scan.width = Math.round(bitmap.width * fit);
  scan.height = Math.round(bitmap.height * fit);
  const scanContext = scan.getContext('2d', { willReadFrequently: true });
  scanContext.imageSmoothingQuality = 'high';
  scanContext.drawImage(bitmap, 0, 0, scan.width, scan.height);
  bitmap.close();

  const box = contentBox(scanContext, scan.width, scan.height);
  const scale = Math.min(SIZE / box.width, SIZE / box.height);
  const canvas = document.createElement('canvas');
  canvas.width = Math.round(box.width * scale);
  canvas.height = Math.round(box.height * scale);
  const context = canvas.getContext('2d');
  context.imageSmoothingQuality = 'high';
  context.drawImage(scan, box.left, box.top, box.width, box.height, 0, 0, canvas.width, canvas.height);
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => (blob ? resolve(blob) : reject(new Error('picture'))), 'image/webp', 0.82);
  });
}

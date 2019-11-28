export function loadImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image();
    image.onload = () => {
      resolve(image);
    };
    image.onerror = (event: any) => {
      const err = new Error(
        `Unable to load image from URL: \`${event.path[0].currentSrc}\``
      );
      reject(err);
    };
    image.crossOrigin = "anonymous";
    image.src = url;
  });
}

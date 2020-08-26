import "../styles/spinner.css";

export function spinnerHTML(): string {
  return [
    '<div class="bounce1"></div>',
    '<div class="bounce2"></div>',
    '<div class="bounce3"></div>',
  ].join("");
}

export function circleSpinnerHTML(): string {
  return '<div class="circle-spinner"></div>';
}

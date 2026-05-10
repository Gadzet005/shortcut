export function formatPrice(price: number): string {
  const whole = Math.round(price).toString();
  const grouped = whole.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  return `${grouped} ₽`;
}

// Fir branches, painted once onto a canvas and never touched again, so all those needles cost nothing afterwards.

const GREENS = ['#0a2e1e', '#0f4029', '#155233', '#1c673f', '#27804e', '#3d9c66'];

/** The same numbers every time, so the garland looks the same after a resize. */
export function seeded(seed) {
  let state = seed;
  return () => {
    state = (state * 16807) % 2147483647;
    return state / 2147483647;
  };
}

/**
 * Covers a line with needles, dark ones first so the light ones lie on top.
 * @param points [{x, y, angle}], the direction the branch grows in at that point
 */
export function bough(ctx, points, { thick, density, length, random }) {
  ctx.lineCap = 'round';
  GREENS.forEach((color, layer) => {
    ctx.strokeStyle = color;
    ctx.lineWidth = 1.5 - layer * 0.1;
    ctx.beginPath();
    for (const point of points) {
      for (let i = 0; i < density; i++) {
        const side = random() < 0.5 ? -1 : 1;
        const spread = (0.35 + random() * 0.85) * side;
        const angle = point.angle + spread;
        const reach = thick * (1 - layer / (GREENS.length + 2)) * random() * 0.5;
        const x = point.x + Math.cos(point.angle + (Math.PI / 2) * side) * reach;
        const y = point.y + Math.sin(point.angle + (Math.PI / 2) * side) * reach;
        const long = length * (0.65 + random() * 0.5);
        ctx.moveTo(x, y);
        ctx.lineTo(x + Math.cos(angle) * long, y + Math.sin(angle) * long);
      }
    }
    ctx.stroke();
  });
}

/** A shiny Christmas ball with the light from the top left. */
export function bauble(ctx, x, y, radius, color, deep) {
  ctx.strokeStyle = '#c9a54e';
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(x, y - radius - 7);
  ctx.lineTo(x, y - radius);
  ctx.stroke();
  ctx.fillStyle = '#d8bd6a';
  ctx.fillRect(x - 2.5, y - radius - 3, 5, 4);
  const shine = ctx.createRadialGradient(x - radius * 0.4, y - radius * 0.4, radius * 0.05, x, y, radius);
  shine.addColorStop(0, '#ffffff');
  shine.addColorStop(0.2, color);
  shine.addColorStop(1, deep);
  ctx.fillStyle = shine;
  ctx.beginPath();
  ctx.arc(x, y, radius, 0, Math.PI * 2);
  ctx.fill();
}

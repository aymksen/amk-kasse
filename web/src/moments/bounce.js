// The path of the logo on the screensaver. Hitting a corner is meant to be rare, so it cannot be left to chance:
// until the next corner is due, a course that would end in one is bent into a near miss, and once it is due, the
// logo aims for the corner ahead at the next wall. All times are in milliseconds, all distances in pixels.
const MINUTE = 60 * 1000;
const CORNER_MIN_MS = 3 * MINUTE;
const CORNER_MAX_MS = 10 * MINUTE;
const NEAR_MISS_PX = 30; // a close call stays this far from the corner
const STEEPEST = 2.2; // about 65 degrees: anything steeper or flatter looks wrong
const FLATTEST = 1 / STEEPEST;

export function nextCorner(now) {
  return now + CORNER_MIN_MS + Math.random() * (CORNER_MAX_MS - CORNER_MIN_MS);
}

function randomSlope() {
  return 0.65 + Math.random() * 0.8;
}

// Same speed and same direction of travel, at a new angle.
function setSlope(p, slope) {
  const across = p.speed / Math.hypot(1, slope);
  p.vx = Math.sign(p.vx) * across;
  p.vy = Math.sign(p.vy) * across * slope;
}

function plan(p, now) {
  const left = { x: p.vx > 0 ? p.maxX - p.x : p.x, y: p.vy > 0 ? p.maxY - p.y : p.y };
  const intoCorner = left.y / left.x;
  if (now >= p.cornerAt && intoCorner >= FLATTEST && intoCorner <= STEEPEST) {
    setSlope(p, intoCorner);
    p.aiming = true;
    return;
  }
  const miss = left.y - left.x * Math.abs(p.vy / p.vx);
  if (Math.abs(miss) < NEAR_MISS_PX) setSlope(p, (left.y - (miss < 0 ? -NEAR_MISS_PX : NEAR_MISS_PX)) / left.x);
}

export function createPath({ maxX, maxY, speed, x, y, cornerAt, now }) {
  const p = { maxX, maxY, speed, x, y, vx: 1, vy: 1, cornerAt, aiming: false };
  setSlope(p, randomSlope());
  plan(p, now);
  return p;
}

export function move(p, seconds, now) {
  p.x += p.vx * seconds;
  p.y += p.vy * seconds;
  let wallX = p.x <= 0 || p.x >= p.maxX;
  let wallY = p.y <= 0 || p.y >= p.maxY;
  if (p.aiming && (wallX || wallY)) wallX = wallY = true; // a rounding error must not cost it the corner
  if (wallX) {
    p.x = p.vx > 0 ? p.maxX : 0;
    p.vx = -p.vx;
  }
  if (wallY) {
    p.y = p.vy > 0 ? p.maxY : 0;
    p.vy = -p.vy;
  }
  const corner = wallX && wallY;
  if (corner) {
    p.aiming = false;
    p.cornerAt = nextCorner(now);
    setSlope(p, randomSlope());
  }
  if (wallX || wallY) plan(p, now);
  return { wallX, wallY, corner };
}

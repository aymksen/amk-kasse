// The drawing the till's monkeys share: the 6,70 € one, the devil and the 4,20 € one. It comes from the monkey in the
// logo. Every number is in units of one 800 × 800 picture, and the right side is the left side mirrored.

export const MIRROR = 'matrix(-1 0 0 1 800 0)';

// A layer is as big as what it holds, so each svg covers only what it draws.
// The numbers are x, y, width and height in the 800 × 800 picture.
export const crop = (x, y, w, h) => ({
  viewBox: `${x} ${y} ${w} ${h}`,
  style: `left:calc(var(--u)*${x});top:calc(var(--u)*${y});width:calc(var(--u)*${w});height:calc(var(--u)*${h})`,
});

// The same box on the other side of the monkey.
export const flip = (x, y, w, h) => crop(800 - x - w, y, w, h);

export const TORSO = 'M330 396C290 406 240 428 226 486C212 544 196 640 196 830H604C604 640 588 544 574 486C560 428 510 406 470 396Z';
export const BELLY = 'M400 546C356 546 322 572 306 622C290 672 292 742 316 790C334 826 368 838 400 838C432 838 466 826 484 790C508 742 510 672 494 622C478 572 444 546 400 546Z';
export const CHEST_TUFT = 'M334 566L350 594L362 566L378 604L392 572L402 610L416 572L430 602L442 566L456 592L470 566C420 530 380 530 334 566Z';

// Both arm parts end in a circle around their joint, so they stay joined at any angle.
export const UPPER_ARM = 'M244 436C205 460 165 500 127 536A38 38 0 0 0 173 596C215 565 265 535 300 508A46 46 0 0 0 244 436Z';
export const FOREARM = 'M184 549C165 515 135 465 112 423A32 32 0 0 0 56 453C78 495 98 540 116 583A38 38 0 0 0 184 549Z';
export const ARM_SHEEN = 'M250 452C212 476 176 508 148 538';
export const FOREARM_SHEEN = 'M86 460C104 494 122 530 138 556';

export const FACE = 'M400 232C384 206 342 196 316 212C288 230 282 272 292 300C288 322 296 362 324 390C350 414 450 414 476 390C504 362 512 322 508 300C518 272 512 230 484 212C458 196 416 206 400 232Z';
export const NOSE = 'M376 322C380 304 420 304 424 322C428 336 410 336 400 330C390 336 372 336 376 322Z';
export const EAR_FOLD = 'M250 310C242 288 262 276 280 288C270 288 262 298 268 312';

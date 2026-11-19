<script>
  import { pictureUrl } from '../lib/picture.js';

  let { product, size = 36 } = $props();

  const url = $derived(pictureUrl(product));
</script>

<span class="thumb" style:width="{size}px" style:height="{size}px">
  {#if url}<img src={url} alt="" width={size} height={size} loading="lazy" decoding="async" />{/if}
</span>

<style>
  .thumb {
    display: inline-block;
    flex-shrink: 0;
    border-radius: 4px;
    background: #fff;
    box-shadow: 0 0 0 1px var(--line);
    overflow: hidden;
    vertical-align: middle;
  }
  img { display: block; width: 100%; height: 100%; object-fit: contain; }
</style>

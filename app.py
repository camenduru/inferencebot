import torch
from diffusers import AutoencoderTiny, StableDiffusionPipeline

from streamdiffusion import StreamDiffusion
from streamdiffusion.image_utils import postprocess_image

pipe = StableDiffusionPipeline.from_pretrained("KBlueLeaf/kohaku-v2.1").to(
    device=torch.device("cuda"),
    dtype=torch.float16,
)

stream = StreamDiffusion(
    pipe,
    t_index_list=[0, 16, 32, 45],
    torch_dtype=torch.float16,
    cfg_type="none",
)

stream.load_lcm_lora()
stream.fuse_lora()
stream.vae = AutoencoderTiny.from_pretrained("madebyollin/taesd").to(device=pipe.device, dtype=pipe.dtype)
pipe.enable_xformers_memory_efficient_attention()

import gradio as gr

def generate(prompt):
  stream.prepare(prompt)
  for _ in range(4):
    stream()
  x_output = stream.txt2img()
  image = postprocess_image(x_output, output_type="pil")[0]
  image.save('/content/image.jpg')
  return image.resize((512, 512))

with gr.Blocks(title=f"Realtime SDXL Turbo", css=".gradio-container {max-width: 544px !important}") as demo:
    with gr.Row():
      with gr.Column():
          textbox = gr.Textbox(show_label=False, value="a close-up picture of a fluffy cat")
          button = gr.Button()
    with gr.Row(variant="default"):
        output_image = gr.Image(
            show_label=False,
            type="pil",
            interactive=False,
            height=512,
            width=512,
            elem_id="output_image",
        )
    button.click(fn=generate, inputs=[textbox], outputs=[output_image], show_progress=False)

demo.queue().launch(inline=False, share=False, debug=False)
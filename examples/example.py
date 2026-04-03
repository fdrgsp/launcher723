# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "ndv[vispy,pyqt]",
#   "imageio[tifffile]"
# ]
# ///

import ndv

data = ndv.data.cells3d()
ndv.imshow(data)
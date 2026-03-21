import math, argparse
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set, Tuple
from scipy.stats import qmc
from opensimplex import OpenSimplex
import numpy as np
import matplotlib.pyplot as plt
import matplotlib.colors as pltc
import matplotlib.patches as mpatches
from matplotlib.patches import RegularPolygon
from matplotlib.collections import PatchCollection

AXIAL_DIRS = [(1, 0), (-1, 0), (0, 1), (0, -1), (1, -1), (-1, 1)]


def hex_neighbors(p: Tuple[int, int]) -> List[Tuple[int, int]]:
    q, r = p
    return [(q + dq, r + dr) for dq, dr in AXIAL_DIRS]


def hex_distance(p1: Tuple[int, int], p2: Tuple[int, int]) -> int:
    """Cube-coordinate distance between two axial hex cells."""
    q1, r1 = p1
    q2, r2 = p2
    return (abs(q2 - q1) + abs(q2 - q1 + r2 - r1) + abs(r2 - r1)) // 2


def axial_to_cart(q: int, r: int) -> Tuple[float, float]:
    """Flat-top hex: x = q + r/2,  y = r * sqrt(3)/2"""
    return q + r * 0.5, r * math.sqrt(3) / 2
    # x = q + (r % 2) * 0.5
    # y = r * math.sqrt(3) / 2
    # return x, y


TRANSLATOR = {
    0: {"label": "ocean", "color": "royalblue"},
    1: {"label": "sea", "color": "cornflowerblue"},
    2: {"label": "beach", "color": "sandybrown"},
    3: {"label": "plains", "color": "forestgreen"},
    4: {"label": "mountain", "color": "darkgray"},
}


@dataclass
class WorldConfig:
    cols: int = 30  # number of columns (q axis)
    rows: int = 20  # number of rows    (r axis)
    border: int = 0
    ncandidates: int = 30
    # per-island tile counts (inclusive ranges)
    m_min: int = 3
    m_max: int = 8
    p_min: int = 9
    p_max: int = 14
    c_min: int = 15
    c_max: int = 20
    noise_scale1: float = 0.01
    noise_scale2: float = 0.10
    noise_scale3: float = 0.30
    # minimum / maximum hex-distance between island centres
    i_min: int = 15
    i_radius: int = 4
    # how many hex rings around land are forced to Sea (not Ocean)
    sea_radius: int = 1
    seed: Optional[int] = None


class WorldGenerator:
    def __init__(self, config: WorldConfig):
        self.config = config
        self.rng = np.random.default_rng()
        self.grid: Dict[Tuple[int, int], int] = {}
        self.cells: List[Tuple[int, int]] = []

    def _in_bounds(self, q: int, r: int) -> bool:
        return 0 <= q < self.config.cols and 0 <= r < self.config.rows

    def _init_grid(self):
        for q in range(self.config.cols):
            for r in range(self.config.rows):
                self.grid[(q, r)] = 0
                self.cells.append((q, r))

    def _generate_islands(self):
        centers = self._generate_island_centers()
        for center in centers:
            land = self._generate_island_shapes(center)
            self._classify_land(land)

    def _generate_island_centers(self) -> None:
        """
        Fast Poisson Disk Sampling in Arbitrary Dimensions

        Args:
            k (int, optional): Samples before rejection. Defaults to 30.
        Returns:
            list[tuple]: Valid center list.
        """
        centers: List[Tuple[int, int]] = []
        engine = qmc.PoissonDisk(
            d=2,
            radius=self.config.i_min
            / max(
                self.config.cols - 2 * self.config.border,
                self.config.rows - 2 * self.config.border,
            ),
            rng=self.rng,
            ncandidates=self.config.ncandidates,
        )
        sample = engine.fill_space()

        for x, y in sample:
            cx = (
                int(x * (self.config.cols - 2 * self.config.border))
                + self.config.border
            )
            cy = (
                int(y * (self.config.rows - 2 * self.config.border))
                + self.config.border
            )
            self.grid[(cx, cy)] = 5
            centers.append((cx, cy))

        return centers

    def _generate_island_shapes(self, center):
        cx, cy = center

        min_size = self.config.m_min + self.config.p_min + self.config.c_min
        max_size = self.config.m_max + self.config.p_max + self.config.c_max

        valid_island = False
        while not valid_island:
            land = set()
            noise = OpenSimplex(seed=int(self.rng.integers(0, 9999)))

            for x in range(max(0, cx - self.config.i_radius), min(self.config.cols, cx + self.config.i_radius)):
                for y in range(max(0, cy - self.config.i_radius), min(self.config.rows, cy + self.config.i_radius)):
                    sx, sy = axial_to_cart(x, y)
                    d = hex_distance((cx, cy), (x, y))

                    n1 = noise.noise2(
                        sx * self.config.noise_scale1, sy * self.config.noise_scale1
                    )
                    n2 = noise.noise2(
                        sx * self.config.noise_scale2, sy * self.config.noise_scale2
                    )
                    n3 = noise.noise2(
                        sx * self.config.noise_scale3, sy * self.config.noise_scale3
                    )
                    nval = n1 + n2 + n3

                    threshold = self.config.i_radius * (1 + 0.4 * nval)
                    if d < threshold:
                        land.add((x, y))

            if min_size <= len(land) <= max_size:
                valid_island = True

        for l in land:
            self.grid[l] = 5

        return land

    def _classify_land(self, land):
        coast_count = 0
        forest_count = 0
        for tile in land:
            is_coast = False
            for neighbor in hex_neighbors(tile):
                if neighbor not in land:
                    is_coast = True
                    self.grid[neighbor] = 1
            if is_coast:
                coast_count += 1
                self.grid[tile] = 2
            else:
                forest_count += 1
                self.grid[tile] = 3

        print(f"Coast: {coast_count}, Forest: {forest_count}")
    
        
        

        
            

    def generate_map(self) -> Dict[Tuple[int, int], int]:
        self._init_grid()
        self._generate_islands()
        # self._grow_land()
        # self._apply_sea_radius()

    def show_map(self) -> None:
        hex_r = 1.0 / math.sqrt(3)

        cvals = list(TRANSLATOR.keys())
        colors = [TRANSLATOR[val]["color"] for val in cvals]
        cmap = pltc.ListedColormap(colors)
        norm = pltc.BoundaryNorm(
            boundaries=[v - 0.5 for v in cvals] + [cvals[-1] + 0.5], ncolors=len(cvals)
        )

        fig, ax = plt.subplots(
            figsize=(10, 8)
        )
        ax.set_aspect("equal")
        ax.set_axis_off()
        fig.patch.set_facecolor("#0d1b2a")

        # Build patches + numeric values
        patches = []
        values = []

        for (q, r), tile in self.grid.items():
            x, y = axial_to_cart(q, r)
            patches.append(
                RegularPolygon(
                    (x, y), numVertices=6, radius=hex_r * 1.00, orientation=0
                )
            )
            values.append(tile)

        collection = PatchCollection(
            patches, cmap=cmap, norm=norm, edgecolor="darkgray", linewidth=0.5
        )
        collection.set_array(values)
        ax.add_collection(collection)
        ax.autoscale_view()

        # cbar = fig.colorbar(collection, ax=ax, ticks=cvals, shrink=0.6)
        # cbar.ax.set_yticklabels([TRANSLATOR[v]["label"] for v in cvals], color="white")
        # cbar.outline.set_edgecolor("#444")
        # cbar.ax.tick_params(colors="white", length=0)

        # Legend
        legend_elements = [
            mpatches.Patch(
                facecolor=TRANSLATOR[t]["color"], label=TRANSLATOR[t]["label"]
            )
            for t in [0, 1, 2, 3, 4]
        ]
        ax.legend(
            handles=legend_elements,
            loc="lower right",
            framealpha=0.85,
            fontsize=8,
            facecolor="#1a1a2e",
            labelcolor="white",
            edgecolor="#444",
        )

        ax.set_title("World Map", color="white", fontsize=12, pad=10)
        fig.tight_layout()
        plt.show()


def parse_args():
    parser = argparse.ArgumentParser(description="Generate a hex map.")
    parser.add_argument("--cols", type=int, default=30)
    parser.add_argument("--rows", type=int, default=20)
    parser.add_argument("--m_min", type=int, default=1)
    parser.add_argument("--m_max", type=int, default=4)
    parser.add_argument("--p_min", type=int, default=2)
    parser.add_argument("--p_max", type=int, default=6)
    parser.add_argument("--c_min", type=int, default=3)
    parser.add_argument("--c_max", type=int, default=8)
    parser.add_argument("--ncandidates", type=int, default=30)
    parser.add_argument("--border", type=int, default=5)
    parser.add_argument("--noise_scale1", type=float, default=0.01)
    parser.add_argument("--noise_scale2", type=float, default=0.1)
    parser.add_argument("--noise_scale3", type=float, default=0.25)
    parser.add_argument("--i_min", type=int, default=6)
    parser.add_argument("--i_radius", type=int, default=4)
    parser.add_argument("--sea_radius", type=int, default=1)
    parser.add_argument("--seed", type=int, default=None)
    return parser.parse_args()


def run():
    config = WorldConfig()
    generator = WorldGenerator(config)
    generator.generate_map()
    generator.show_map()


if __name__ == "__main__":
    run()

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


TRANSLATOR = {
    0: {"label": "ocean", "color": "royalblue"},
    1: {"label": "sea", "color": "cornflowerblue"},
    2: {"label": "beach", "color": "sandybrown"},
    3: {"label": "plains", "color": "forestgreen"},
    4: {"label": "mountain", "color": "dimgray"},
}


@dataclass
class WorldConfig:
    """World configurations class"""
    cols: int = 100  # number of columns (q axis)
    rows: int = 100  # number of rows    (r axis)
    border: int = 1
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
    # minimum hex-distance between island centres
    i_dist: int = 15
    i_radius: int = 4
    seed: Optional[int] = None


class WorldGenerator:
    """World generator class"""

    def __init__(self, config: WorldConfig):
        """Initialize the world generator.

        Args:
            config (WorldConfig): World configurations.
        """
        self.config = config
        self.rng = np.random.default_rng()
        self.grid: Dict[Tuple[int, int], int] = {}

    def _init_grid(self) -> None:
        """Initialize the grid with ocean tiles.

        Returns:
            None
        """
        for q in range(self.config.cols):
            for r in range(self.config.rows):
                self.grid[(q, r)] = 0

    def _generate_islands(self) -> None:
        """Generate larger islands on the grid.

        Returns:
            None
        """
        centers = self._generate_island_centers()
        for center in centers:
            island = self._generate_island_shapes(center)
            self._classify_island(island)

    def _generate_island_centers(self) -> List[Tuple[int, int]]:
        """Generate island centers using Poisson Disk Sampling.

        Returns:
            list[tuple]: Valid center list.
        """
        search_width = self.config.cols - 2 * self.config.border
        search_height = self.config.rows - 2 * self.config.border

        centers: List[Tuple[int, int]] = []
        engine = qmc.PoissonDisk(
            d=2,
            radius=self.config.i_dist / max(search_width, search_height),
            rng=self.rng,
            ncandidates=self.config.ncandidates,
        )
        sample = engine.fill_space()

        for x, y in sample:
            cx = int(x * search_width) + self.config.border
            cy = int(y * search_height) + self.config.border
            self.grid[(cx, cy)] = 5
            centers.append((cx, cy))

        return centers

    def _generate_island_shapes(self, center) -> Set[Tuple[int, int]]:
        """Generate island shape using Perlin noise.

        Args:
            center (Tuple[int, int]): Center of the island.

        Returns:
            Set[Tuple[int, int]]: Set of hexes that make up the island.
        """
        cx, cy = center

        valid_island = False
        while not valid_island:
            island = set()
            noise = OpenSimplex(seed=int(self.rng.integers(0, 9999)))

            for x in range(
                max(0, cx - self.config.i_radius),
                min(self.config.cols, cx + self.config.i_radius),
            ):
                for y in range(
                    max(0, cy - self.config.i_radius),
                    min(self.config.rows, cy + self.config.i_radius),
                ):
                    d = hex_distance((cx, cy), (x, y))

                    n1 = noise.noise2(
                        x * self.config.noise_scale1, y * self.config.noise_scale1
                    )
                    n2 = noise.noise2(
                        x * self.config.noise_scale2, y * self.config.noise_scale2
                    )
                    n3 = noise.noise2(
                        x * self.config.noise_scale3, y * self.config.noise_scale3
                    )
                    noise_value = n1 + n2 + n3

                    threshold = self.config.i_radius * (1 + 0.4 * noise_value)
                    if d < threshold:
                        island.add((x, y))

            sea, coast, plains, mountains = self._get_biomes(island)
            if self.config.c_min <= len(coast) <= self.config.c_max:
                if self.config.p_min <= len(plains) <= self.config.p_max:
                    if self.config.m_min <= len(mountains) <= self.config.m_max:
                        valid_island = True

        return island

    def _get_biomes(self, island) -> Tuple[
        Set[Tuple[int, int]],
        Set[Tuple[int, int]],
        Set[Tuple[int, int]],
        Set[Tuple[int, int]],
    ]:
        """Separate island into biomes (sea, coast, plains, mountains).

        Args:
            island (Set[Tuple[int, int]]): Set of hexes that make up the island.

        Returns:
            Tuple[Set[Tuple[int, int]], Set[Tuple[int, int]], Set[Tuple[int, int]], Set[Tuple[int, int]]]:
            Set of hexes that make up the sea, coast, plains, and mountains.
        """
        sea = set()
        coast = set()
        plains_mountains = set()
        for tile in island:
            is_coast = False
            for neighbor in hex_neighbors(tile):
                if neighbor not in island:
                    sea.add(neighbor)
                    is_coast = True
            if is_coast:
                coast.add(tile)
            else:
                plains_mountains.add(tile)

        plains = set()
        mountains = set()
        for tile in plains_mountains:
            all_neighbors_plains = True
            for neighbor in hex_neighbors(tile):
                if neighbor not in plains_mountains:
                    all_neighbors_plains = False
                    break
            if all_neighbors_plains:
                mountains.add(tile)
            else:
                plains.add(tile)

        return sea, coast, plains, mountains

    def _classify_island(self, island) -> None:
        """Classify island into biomes.

        Args:
            island (Set[Tuple[int, int]]): Set of hexes that make up the island.

        Returns:
            None
        """
        sea, coast, plains, mountains = self._get_biomes(island)

        for tile in sea:
            self.grid[tile] = 1
        for tile in coast:
            self.grid[tile] = 2
        for tile in plains:
            self.grid[tile] = 3
        for tile in mountains:
            self.grid[tile] = 4

    def generate_map(self) -> Dict[Tuple[int, int], int]:
        """Generate the map.

        Returns:
            Dict[Tuple[int, int], int]: Dictionary of hexes and their types.
        """
        self._init_grid()
        self._generate_islands()
        return self.grid

    def show_map(self) -> None:
        """Show the map.

        Returns:
            None
        """
        hex_r = 1.0 / math.sqrt(3)

        cvals = list(TRANSLATOR.keys())
        colors = [TRANSLATOR[val]["color"] for val in cvals]
        cmap = pltc.ListedColormap(colors)
        norm = pltc.BoundaryNorm(
            boundaries=[v - 0.5 for v in cvals] + [cvals[-1] + 0.5], ncolors=len(cvals)
        )

        fig, ax = plt.subplots(figsize=(10, 8))
        ax.set_aspect("equal")
        ax.set_axis_off()
        fig.patch.set_facecolor("#0d1b2a")

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
            patches, cmap=cmap, norm=norm, edgecolor="darkgray", linewidth=0.0
        )
        collection.set_array(values)
        ax.add_collection(collection)
        ax.autoscale_view()

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


def parse_args() -> argparse.Namespace:
    """Parse command line arguments.

    Returns:
        argparse.Namespace: Parsed arguments.
    """
    parser = argparse.ArgumentParser(description="Generate a hex map.")
    parser.add_argument("--cols", type=int, default=50)
    parser.add_argument("--rows", type=int, default=50)
    parser.add_argument("--border", type=int, default=5)
    parser.add_argument("--ncandidates", type=int, default=30)
    parser.add_argument("--m_min", type=int, default=3)
    parser.add_argument("--m_max", type=int, default=8)
    parser.add_argument("--p_min", type=int, default=9)
    parser.add_argument("--p_max", type=int, default=14)
    parser.add_argument("--c_min", type=int, default=15)
    parser.add_argument("--c_max", type=int, default=20)
    parser.add_argument("--noise_scale1", type=float, default=0.01)
    parser.add_argument("--noise_scale2", type=float, default=0.1)
    parser.add_argument("--noise_scale3", type=float, default=0.25)
    parser.add_argument("--i_dist", type=int, default=15)
    parser.add_argument("--i_radius", type=int, default=4)
    parser.add_argument("--seed", type=int, default=None)
    return parser.parse_args()


def run() -> None:
    """Run the world generator.

    Returns:
        None
    """
    args = parse_args()
    config = WorldConfig(**vars(args))
    generator = WorldGenerator(config)
    generator.generate_map()
    generator.show_map()


if __name__ == "__main__":
    run()

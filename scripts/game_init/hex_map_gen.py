import math
import random
import os
import sys, json
from typing import Dict, List, Set, Tuple

import matplotlib.colors as pltc
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt
import numpy as np
import yaml
from matplotlib.collections import PatchCollection
from matplotlib.patches import RegularPolygon
from opensimplex import OpenSimplex
from pydantic.dataclasses import dataclass
from scipy.stats import qmc

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

    name: str = "world"
    size: int = 256
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


class WorldGenerator:
    """World generator class"""

    def __init__(self, config: WorldConfig) -> None:
        """Initialize the world generator.

        Args:
            config (WorldConfig): World configurations.
        """
        self.config = config
        self.rng = np.random.default_rng()
        self.grid: Dict[Tuple[int, int], int] = {}
        self.settleable: List[Tuple[int, int]] = []

    def _init_grid(self) -> None:
        """Initialize the grid with ocean tiles.

        Returns:
            None
        """
        for q in range(self.config.size):
            for r in range(self.config.size):
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
        search_width = self.config.size - 2 * self.config.border
        search_height = self.config.size - 2 * self.config.border

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

    def _generate_island_shapes(self, center: Tuple[int, int]) -> Set[Tuple[int, int]]:
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
                min(self.config.size, cx + self.config.i_radius),
            ):
                for y in range(
                    max(0, cy - self.config.i_radius),
                    min(self.config.size, cy + self.config.i_radius),
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

    def _get_biomes(self, island: Set[Tuple[int, int]]) -> Tuple[
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

    def _classify_island(self, island: Set[Tuple[int, int]]) -> None:
        """Classify island into biomes and update settleable tiles.

        Args:
            island (Set[Tuple[int, int]]): Set of hexes that make up the island.

        Returns:
            None
        """
        sea, coast, plains, mountains = self._get_biomes(island)

        self._get_settleable(coast, plains, mountains)

        for tile in sea:
            self.grid[tile] = 1
        for tile in coast:
            self.grid[tile] = 2

        for tile in plains:
            self.grid[tile] = 3
        for tile in mountains:
            self.grid[tile] = 4

        

    def _get_settleable(self, coast: Set[Tuple[int, int]], plains: Set[Tuple[int, int]], mountains: Set[Tuple[int, int]]) -> None:
        """Get settleable tiles.

        Returns:
            None
        """
        settleable = set()
        c_cities = random.sample(list(coast), k=self.config.c_min)
        p_cities = random.sample(list(plains), k=self.config.p_min)
        m_cities = random.sample(list(mountains), k=self.config.m_min)

        for city in c_cities:
            settleable.add(city)
        for city in p_cities:
            settleable.add(city)
        for city in m_cities:
            settleable.add(city)

        self.settleable.extend(settleable)

    def generate_map(self) -> Dict[Tuple[int, int], int]:
        """Generate the map.

        Returns:
            Dict[Tuple[int, int], int]: Dictionary of hexes and their types.
        """
        self._init_grid()
        self._generate_islands()
        return self.grid

    def generate_image(self, show: bool = True) -> None:
        """Generate the map image.

        Args:
            show (bool): Whether to show the map.

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

        # add cities placeholders
        city_patches = []
        for q, r in self.settleable:
            x, y = axial_to_cart(q, r)
            city_patches.append(mpatches.Circle((x, y), radius=hex_r * 0.4))

        if city_patches:
            city_collection = PatchCollection(
                city_patches, facecolor="gray", edgecolor="none", zorder=2
            )
            ax.add_collection(city_collection)

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
        ax.autoscale_view()

        ax.set_title("World Map", color="white", fontsize=12, pad=10)
        fig.tight_layout()

        os.makedirs("./world_data", exist_ok=True)
        fig.savefig(
            f"./world_data/{self.config.name}.png",
            bbox_inches="tight",
            facecolor=fig.get_facecolor(),
            edgecolor="none",
        )
        if show:
            plt.show()
        plt.close()

    def save_to_file(self) -> None:
        """Save the map to a file.

        Returns:
            None
        """
        lines = []
        for q in range(self.config.size):
            line = ""
            for r in range(self.config.size):
                line += f"{self.grid[(q, r)]},"
            line += "\n"
            lines.append(line)

        os.makedirs("./world_data", exist_ok=True)

        with open(os.path.join("world_data", f"{self.config.name}.csv"), "w") as f:
            f.writelines(lines)

        with open(os.path.join("world_data", f"{self.config.name}_settleable.json"), "w") as f:
            json.dump(self.settleable, f)

        print(f"✅ Map successfully saved at ./world_data/{self.config.name}.csv")
        print(f"✅ Settleable tiles saved at ./world_data/{self.config.name}_settleable.json")


def load_configs() -> WorldConfig:
    if len(sys.argv) != 2:
        raise ValueError("⛔ Usage: python hex_map_gen.py <config_file>")

    filename = sys.argv[1] + ".yaml" if ".yaml" not in sys.argv[1] else sys.argv[1]

    with open(filename) as f:
        data = yaml.safe_load(f)

    return WorldConfig(**data)


def validate_config(config: WorldConfig) -> None:
    """Validate the config.

    Args:
        config (WorldConfig): The config to validate.

    Returns:
        None
    """
    if config.size <= 0:
        raise ValueError("⛔ Size must be positive")
    elif config.size <= (config.i_radius + config.border) * 2:
        raise ValueError(
            "⛔ Size must be larger than twice the island radius plus border"
        )

    if config.border < 0:
        raise ValueError("⛔ Border must be non-negative")
    elif config.border >= config.size / 2:
        raise ValueError("⛔ Border must be less than half the world size")
    elif config.border <= config.i_radius:
        print("⚠️ Border is too small, may result in islands over the border")

    if config.ncandidates <= 0:
        raise ValueError(
            "⛔ Number of candidates must be positive. Suggested value: 30"
        )
    elif config.ncandidates > config.size * config.size:
        raise ValueError("⛔ Number of candidates must be less than the world size")
    elif config.ncandidates > 60:
        print("⚠️ Number of candidates is large, may result long generation time")

    if (
        config.m_min < 0
        or config.m_max < 0
        or config.p_min < 0
        or config.p_max < 0
        or config.c_min < 0
        or config.c_max < 0
    ):
        raise ValueError("⛔ Tile counts must be non-negative")
    elif (
        config.m_min > config.m_max
        or config.p_min > config.p_max
        or config.c_min > config.c_max
    ):
        raise ValueError(
            "⛔ Minimum tile count must be less than or equal to maximum tile count"
        )
    elif config.m_max + config.p_max + config.c_max > config.size * config.size:
        raise ValueError("⛔ Maximum tile count must be less than the world size")

    if config.noise_scale1 <= 0 or config.noise_scale2 <= 0 or config.noise_scale3 <= 0:
        raise ValueError("⛔ Noise scales must be positive")
    elif config.noise_scale1 > 1 or config.noise_scale2 > 1 or config.noise_scale3 > 1:
        print("⚠️ Noise scales are large, may result in too noisy islands")

    if config.i_dist <= 0 or config.i_radius <= 0:
        raise ValueError("⛔ Island distance and radius must be positive")
    elif config.i_dist < config.i_radius * 2:
        raise ValueError(
            "⛔ Island distance must be greater than twice the island radius"
        )
    elif config.i_dist > config.size / 2:
        raise ValueError("⛔ Island distance must be less than half the world size")


def run() -> None:
    """Run the world generator.

    Returns:
        None
    """
    config = load_configs()
    validate_config(config)
    generator = WorldGenerator(config)
    generator.generate_map()
    generator.generate_image()
    generator.save_to_file()


if __name__ == "__main__":
    run()

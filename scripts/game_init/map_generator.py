import random, math
import numpy as np
import matplotlib.pyplot as plt
import matplotlib.colors as pltc
from scipy.stats import qmc
from matplotlib.collections import PatchCollection

def norm1_distance(a, b):
    return abs(a[0] - b[0]) + abs(a[1] - b[1])


def norm2_distance(a, b):
    return math.sqrt((a[0] - b[0]) ** 2 + (a[1] - b[1]) ** 2)


class WorldGenerator:
    world_w = 256
    world_h = 256
    i_min = 10
    isl_size = 7
    m_cities = 3
    max_mountain = 5
    p_cities = 6
    max_plains = 3
    c_cities = 6
    max_coast = 3

    def __init__(self):
        self.world = [[0 for _ in range(self.world_w)] for _ in range(self.world_h)]

    def generate_map(self):
        self.centers = self.generate_island_centers()
        for center in self.centers:
            x, y = center
            print(x, y)
            self.world[x][y] = 5

    def generate_island_centers(self, k=30) -> list[tuple]:
        """
        Fast Poisson Disk Sampling in Arbitrary Dimensions

        Args:
            k (int, optional): Samples before rejection. Defaults to 30.

        Returns:
            list[tuple]: Valid center list.
        """

        rng = np.random.default_rng()
        engine = qmc.PoissonDisk(
            d=2, radius=self.i_min / max(self.world_w, self.world_h), rng=rng
        )
        sample = engine.fill_space()

        centers = []
        for x, y in sample:
            cx = int(x * self.world_w)
            cy = int(y * self.world_h)
            centers.append((cx, cy))

        return centers

    def generate_island(self):
        pass

    def show_map(self):
        cvals = [1, 2, 3, 4, 5]
        colors = [
            "royalblue",
            "cornflowerblue",
            "navajowhite",
            "forestgreen",
            "dimgrey",
        ]

        fig, ax = plt.subplots()

        norm = plt.Normalize(min(cvals), max(cvals))
        tuples = list(zip(map(norm, cvals), colors))
        cmap = pltc.LinearSegmentedColormap.from_list("", tuples)
        cmap = pltc.ListedColormap(colors)

        for xi, yi in self.centers:
            ax.Circle((xi, yi), radius=self.i_min / max(self.world_w, self.world_h)/2, fill=False)

        # Plot with matplotlib
        ax.imshow(self.world, cmap=cmap, vmin=0.5, vmax=5.5)
        ax.colorbar()
        ax.title("World Map")
        plt.show()

    def run(self):
        self.generate_map()
        self.show_map()


def run():
    generator = WorldGenerator()
    generator.run()


if __name__ == "__main__":
    run()
